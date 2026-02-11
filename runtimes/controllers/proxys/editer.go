package proxys

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/proxys"
	"tools/runtimes/db/tag"
	"tools/runtimes/eventbus"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/logs"
	"tools/runtimes/parses"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	// tx := db.DB.Begin()
	if err := db.DB.Write(func(tx *gorm.DB) error {
		if id != "" {
			idi, err := strconv.Atoi(id)
			if err != nil {
				// response.Error(c, http.StatusNotFound, i18n.T("Agent does not exist"), nil)
				return errors.New(i18n.T("Agent does not exist"))
			}
			dbid = int64(idi)
			px = proxys.GetById(dbid)
			if px.Id < 1 {
				// response.Error(c, http.StatusNotFound, i18n.T("Agent does not exist"), nil)
				return errors.New(i18n.T("Agent does not exist"))
			}
			if px.Subscribe > 0 {
				// response.Error(c, http.StatusNotFound, i18n.T("Subscribed agents cannot be modified manually"), nil)
				return errors.New(i18n.T("Subscribed agents cannot be modified manually"))
			}
		}

		if pcr.Port != 0 && pcr.Port < config.PROXYMINPORT {
			// response.Error(c, http.StatusNotFound, i18n.T("The port number cannot be less than %d", config.PROXYMINPORT), nil)
			return errors.New(i18n.T("The port number cannot be less than %d", config.PROXYMINPORT))
		}

		// 检查端口是否重复, 要排除修改代理时被修改的代理端口和自身的端口一致问题
		if pcr.Port > 0 {
			portProxy := proxys.GetByPort(pcr.Port)
			if portProxy != nil && portProxy.Id > 0 {
				if dbid > 0 && portProxy.Id != dbid {
					// response.Error(c, http.StatusBadRequest, i18n.T("Port %d is already in use", pcr.Port), nil)
					return errors.New(i18n.T("Port %d is already in use", pcr.Port))
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

		if err := px.Save(px, tx); err != nil {
			// tx.Rollback()
			// response.Error(c, http.StatusNotFound, err.Error(), nil)
			return err
		}

		if len(pcr.Tags) > 0 {
			if err := px.CoverTgs(pcr.Tags, tx); err != nil {
				logs.Error(err.Error())
				// response.Error(c, http.StatusNotFound, err.Error(), nil)
				return err
			}
		}

		if needGetLocal == true {
			if loc, err := proxy.GetLocal(px.GetConfig(), px.GetTransfer()); err == nil {
				px.Local = loc.Iso
				px.Lang = loc.Lang
				px.Timezone = loc.Timezone
				px.Ip = loc.Ip
				if err := px.Save(px, tx); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	// 如果代理已经是启动状态,需要重新启动
	if pcc := px.IsStart(); pcc != nil {
		pcc.Restart(px.Port)
	} else if px.AutoRun == 1 {
		if pcc, err := proxy.Client(px.GetConfig(), px.ListerAddr, px.Port, px.GetTransfer()); err == nil && pcc != nil {
			pcc.Run(true)
		}
	}

	pxc, _ := parses.Marshal(px, c)
	response.Success(c, pxc, "Success")
}

// 删除代理
func Remove(c *gin.Context) {
	id := c.Param("id")
	pc := proxys.GetById(id)
	if pc != nil && pc.Id > 0 {
		if client, err := proxy.Client(pc.GetConfig(), "", 0); err == nil {
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
		// tx := db.DB.Begin()
		if err := db.DB.Write(func(tx *gorm.DB) error {
			if err := tx.Where("id in ?", ids.Ids).Delete(&proxys.Proxy{}).Error; err != nil {
				return err
			}
			if err := tx.Where("proxy_id in ?", ids.Ids).Delete(&proxys.ProxyTag{}).Error; err != nil {
				// tx.Rollback()
				// response.Error(c, http.StatusBadGateway, "", nil)
				return err
			}
			return nil
		}); err != nil {
			response.Error(c, http.StatusBadGateway, err.Error(), nil)
			return
		}
		// if err := tx.Where("id in ?", ids.Ids).Delete(&proxys.Proxy{}).Error; err != nil {
		// 	tx.Rollback()
		// 	response.Error(c, http.StatusBadGateway, "", nil)
		// 	return
		// }
		// if err := tx.Where("proxy_id in ?", ids.Ids).Delete(&proxys.ProxyTag{}).Error; err != nil {
		// 	tx.Rollback()
		// 	response.Error(c, http.StatusBadGateway, "", nil)
		// 	return
		// }
		// tx.Commit()
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
	db.DB.DB().Model(&proxys.Proxy{}).Where("config_md5 in ?", ccmd5).Find(&dbhads)
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
		if err := db.DB.Write(func(tx *gorm.DB) error {
			return tx.Create(pxs).Error
		}); err == nil {
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
					db.DB.Write(func(tx *gorm.DB) error {
						return tx.Create(ptags).Error
					})
				}
			}

			for _, v := range pxs {
				go func() {
					if loc, err := proxy.GetLocal(v.GetConfig(), v.GetTransfer()); err == nil {
						v.Local = loc.Iso
						v.Timezone = loc.Timezone
						v.Lang = loc.Lang
						v.Ip = loc.Ip
						db.DB.Write(func(tx *gorm.DB) error {
							return v.Save(v, tx)
						})
					}
				}()
			}
		}
	}
	response.Success(c, nil, "")
}

func Ping(c *gin.Context) {
	ids := c.Param("id")
	if ids == "" {
		response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
		return
	}

	ur, ok := c.Get("_user")
	if !ok {
		response.Error(c, http.StatusBadRequest, i18n.T("Login first"), nil)
		return
	}
	user := ur.(*admins.Admin)
	// rspmp := make(map[int64]map[string]int64)
	for _, id := range strings.Split(ids, ",") {
		pc := proxys.GetById(id)
		if pc != nil && pc.Id > 0 {
			pxc, err := proxy.Client(pc.GetConfig(), "", 0)
			if err != nil {
				response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
				return
			}

			go func(uid, pid int64) {
				// proxy-ping
				mmp, err := pxc.Delay([]string{"https://www.google.com"})
				pr := proxys.PingResp{UID: uid}
				pr.Ping = make(map[int64]int64)
				if err != nil {
					// response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
					// return
					pr.Ping[pid] = -1
				} else {
					for _, v := range mmp {
						pr.Ping[pid] = v
						break
					}
				}
				eventbus.Bus.Publish("proxy-ping", pr)
			}(user.Id, pc.Id)
			// mmp, err := pxc.Delay([]string{"https://www.google.com"})
			// if err != nil {
			// 	response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
			// 	return
			// }
			// response.Success(c, mmp, "")
			// return
			// rspmp[pc.Id] = mmp
		}
	}
	response.Success(c, nil, "success")
}

// 批量修改
type batchEditeData struct {
	Ids      []int64  `json:"ids" form:"ids"`
	Username string   `json:"username" form:"username"`
	Password string   `json:"password" form:"password"`
	Remark   string   `json:"remark" form:"remark"`
	Transfer string   `json:"transfer" form:"transfer"`
	Tags     []string `json:"tags" form:"tags"`
}

func BatchEditer(c *gin.Context) {
	dt := new(batchEditeData)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	needUpdate := false
	taglen := 0
	var uptags map[string]int64
	upto := make(map[string]interface{})
	if dt.Username != "" {
		upto["username"] = dt.Username
		needUpdate = true
	}
	if dt.Password != "" {
		upto["password"] = dt.Password
		needUpdate = true
	}
	if dt.Remark != "" {
		upto["remark"] = dt.Remark
		needUpdate = true
	}
	if dt.Transfer != "" {
		upto["transfer"] = dt.Transfer
		needUpdate = true
	}
	if len(dt.Tags) > 0 {
		uptags = tag.GetTagsByNames(dt.Tags, nil)
		for _, _ = range uptags {
			taglen++
		}
	}

	// tx := db.DB.Begin()
	if err := db.DB.Write(func(tx *gorm.DB) error {
		if needUpdate == true {
			if err := tx.Model(&proxys.Proxy{}).Where("id in ?", dt.Ids).Updates(upto).Error; err != nil {
				return err
			}
		}

		if uptags != nil && taglen > 0 {
			if err := tx.Where("proxy_id in ?", dt.Ids).Delete(&proxys.ProxyTag{}).Error; err != nil {
				return err
			}

			var dbProxyTag []*proxys.ProxyTag
			for _, pid := range dt.Ids {
				for _, tagid := range uptags {
					dbProxyTag = append(dbProxyTag, &proxys.ProxyTag{
						ProxyId: pid,
						TagId:   tagid,
					})
				}
			}
			if len(dbProxyTag) > 0 {
				if err := tx.Create(dbProxyTag).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	// if needUpdate == true {
	// 	if err := tx.Model(&proxys.Proxy{}).Where("id in ?", dt.Ids).Updates(upto).Error; err != nil {
	// 		tx.Rollback()
	// 		response.Error(c, http.StatusBadGateway, err.Error(), nil)
	// 		return
	// 	}
	// }

	// if uptags != nil && taglen > 0 {
	// 	if err := tx.Where("proxy_id in ?", dt.Ids).Delete(&proxys.ProxyTag{}).Error; err != nil {
	// 		tx.Rollback()
	// 		response.Error(c, http.StatusBadGateway, err.Error(), nil)
	// 		return
	// 	}

	// 	var dbProxyTag []*proxys.ProxyTag
	// 	for _, pid := range dt.Ids {
	// 		for _, tagid := range uptags {
	// 			dbProxyTag = append(dbProxyTag, &proxys.ProxyTag{
	// 				ProxyId: pid,
	// 				TagId:   tagid,
	// 			})
	// 		}
	// 	}
	// 	if len(dbProxyTag) > 0 {
	// 		if err := tx.Create(dbProxyTag).Error; err != nil {
	// 			tx.Rollback()
	// 			response.Error(c, http.StatusBadGateway, err.Error(), nil)
	// 			return
	// 		}
	// 	}
	// }

	// tx.Commit()
	response.Success(c, nil, "")
}
