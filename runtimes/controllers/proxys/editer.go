package proxys

import (
	"net/http"
	"strconv"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/proxys"
	"tools/runtimes/db/tag"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/logs"
	"tools/runtimes/parses"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func Editer(c *gin.Context) {
	pcr := new(proxys.Proxy)
	if err := c.ShouldBindJSON(pcr); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	if pcr.Config == "" {
		response.Error(c, http.StatusNotFound, i18n.T("Configuration cannot be empty"), nil)
		return
	}

	var dbid int64
	id := c.Param("id")
	px := new(proxys.Proxy)

	tx := db.DB.Begin()

	if id != "" {
		idi, err := strconv.Atoi(id)
		if err != nil {
			response.Error(c, http.StatusNotFound, i18n.T("Agent does not exist"), nil)
			return
		}
		dbid = int64(idi)
		px = proxys.GetById(dbid)
		if px.Id < 1 {
			response.Error(c, http.StatusNotFound, i18n.T("Agent does not exist"), nil)
			return
		}
		if px.Subscribe > 0 {
			response.Error(c, http.StatusNotFound, i18n.T("Subscribed agents cannot be modified manually"), nil)
			return
		}
	}

	if pcr.Port != 0 && pcr.Port < config.PROXYMINPORT {
		response.Error(c, http.StatusNotFound, i18n.T("The port number cannot be less than %d", config.PROXYMINPORT), nil)
		return
	}

	// 检查端口是否重复, 要排除修改代理时被修改的代理端口和自身的端口一致问题
	if pcr.Port > 0 {
		portProxy := proxys.GetByPort(pcr.Port)
		if portProxy != nil && portProxy.Id > 0 {
			if dbid > 0 && portProxy.Id != dbid {
				response.Error(c, http.StatusBadRequest, i18n.T("Port %d is already in use", pcr.Port), nil)
				return
			}
		}
	}

	needGetLocal := false
	if px.Id > 0 {
		px.Name = pcr.Name
		if pcr.Config != px.Config {
			px.Config = pcr.Config
			px.ConfigMd5 = funcs.Md5String(pcr.Config)
			needGetLocal = true
		}
		px.AutoRun = pcr.AutoRun
		px.Password = pcr.Password
		px.Port = pcr.Port
		px.Remark = pcr.Remark
		px.Transfer = pcr.Transfer
		px.Username = pcr.Username
	} else {
		needGetLocal = true

		px.AutoRun = pcr.AutoRun
		px.Name = pcr.Name
		px.Config = pcr.Config
		px.ConfigMd5 = funcs.Md5String(pcr.Config)
		px.Password = pcr.Password
		px.Port = pcr.Port
		px.Remark = pcr.Remark
		px.Transfer = pcr.Transfer
		px.Username = pcr.Username
	}

	if err := px.Save(tx); err != nil {
		tx.Rollback()
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	if len(pcr.Tags) > 0 {
		if err := px.CoverTgs(pcr.Tags, tx); err != nil {
			logs.Error(err.Error())
			tx.Rollback()
			response.Error(c, http.StatusNotFound, err.Error(), nil)
			return
		}
	}

	if needGetLocal == true {
		if loc, err := proxy.GetLocal(px.Config, px.Transfer); err == nil {
			px.Local = loc
			px.Save(tx)
		}
	}

	tx.Commit()

	// 如果代理已经是启动状态,需要重新启动
	if pcc := px.IsStart(); pcc != nil {
		pcc.Restart(px.Port)
	}

	pxc, _ := parses.Marshal(px, c)
	response.Success(c, pxc, "Success")
}

// 删除代理
func Remove(c *gin.Context) {
	id := c.Param("id")
	pc := proxys.GetById(id)
	if pc != nil && pc.Id > 0 {
		if client, err := proxy.Client(pc.Config, "", 0); err == nil {
			if client != nil && client.IsRuning() {
				client.Close(true)
			}
		}
	}
	pc.Remove()
	response.Success(c, nil, "")
}

// 批量删除
type rmvsdt struct {
	Ids []int64 `json:"ids" form:"ids"`
}

func Removes(c *gin.Context) {
	ids := new(rmvsdt)
	if err := c.ShouldBind(ids); err == nil {
		tx := db.DB.Begin()
		if err := tx.Where("id in ?", ids.Ids).Delete(&proxys.Proxy{}).Error; err != nil {
			tx.Rollback()
			response.Error(c, http.StatusBadGateway, "", nil)
			return
		}
		if err := tx.Where("proxy_id in ?", ids.Ids).Delete(&proxys.ProxyTag{}).Error; err != nil {
			tx.Rollback()
			response.Error(c, http.StatusBadGateway, "", nil)
			return
		}
		tx.Commit()
	}
	response.Success(c, nil, "")
}

// 批量添加代理
func BatchAdd(c *gin.Context) {
	pcr := new(proxys.Proxy)
	if err := c.ShouldBindJSON(pcr); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	if pcr.Id > 0 {
		response.Error(c, http.StatusNotFound, i18n.T("Batch editing is not possible"), nil)
		return
	}

	if pcr.Config == "" {
		response.Error(c, http.StatusNotFound, i18n.T("Configuration cannot be empty"), nil)
		return
	}

	ccs := strings.ReplaceAll(pcr.Config, "\r\n", "\n")
	ccs = strings.ReplaceAll(ccs, "\r", "\n")
	configs := strings.Split(ccs, "\n")

	var ccmd5 []string
	for _, v := range configs {
		ccmd5 = append(ccmd5, funcs.Md5String(v))
	}

	var dbhads []*proxys.Proxy
	db.DB.Model(&proxys.Proxy{}).Where("config_md5 in ?", ccmd5).Find(&dbhads)
	dbhadMap := make(map[string]byte)
	for _, v := range dbhads {
		dbhadMap[v.ConfigMd5] = 1
	}

	var pxs []*proxys.Proxy
	for _, v := range configs {
		if strings.Trim(v, " ") == "" {
			continue
		}
		cmd5 := funcs.Md5String(v)
		if _, ok := dbhadMap[cmd5]; ok {
			continue
		}

		ot, err := proxy.Client(v, "", 0)
		if err != nil || ot == nil {
			continue
		}

		name := ot.Name
		if name == "" {
			name = funcs.RoundmUuid()
		}

		px := new(proxys.Proxy)
		px.AutoRun = pcr.AutoRun
		px.Name = name
		px.Config = v
		px.ConfigMd5 = cmd5
		px.Remark = pcr.Remark
		px.Transfer = pcr.Transfer
		px.Username = pcr.Username
		px.Password = pcr.Password

		pxs = append(pxs, px)
	}

	if len(pxs) > 0 {
		if err := db.DB.Create(pxs).Error; err == nil {
			if len(pcr.Tags) > 0 {
				tagmp := tag.GetTagsByNames(pcr.Tags, nil)
				var tagids []int64
				for _, v := range pcr.Tags {
					if tid, ok := tagmp[v]; ok {
						tagids = append(tagids, tid)
					}
				}

				var ptags []*proxys.ProxyTag
				for _, v := range pxs {
					for _, tagid := range tagids {
						ttt := new(proxys.ProxyTag)
						ttt.ProxyId = v.Id
						ttt.TagId = tagid
						ptags = append(ptags, ttt)
					}
				}

				if len(ptags) > 0 {
					db.DB.Create(ptags)
				}
			}

			for _, v := range pxs {
				go func() {
					if loc, err := proxy.GetLocal(v.Config, v.Transfer); err == nil {
						v.Local = loc
						v.Save(nil)
					}
				}()
			}
		}
	}
	response.Success(c, nil, "")
}

// 延迟
func Ping(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
		return
	}
	pc := proxys.GetById(id)
	if pc != nil && pc.Id > 0 {
		pxc, err := proxy.Client(pc.Config, "", 0)
		if err != nil {
			response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
			return
		}
		mmp, err := pxc.Delay([]string{"https://www.google.com"})
		if err != nil {
			response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
			return
		}
		response.Success(c, mmp, "")
		return
	}
	response.Error(c, http.StatusBadRequest, "", nil)
}
