package proxys

import (
	"net/http"
	"strconv"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/proxys"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func Editer(c *gin.Context) {
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

	pcr := new(proxys.Proxy)
	if err := c.ShouldBindJSON(pcr); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	if pcr.Config == "" {
		response.Error(c, http.StatusNotFound, i18n.T("Configuration cannot be empty"), nil)
		return
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
