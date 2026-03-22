package users

import (
	"fmt"
	"math/rand"
	"net/http"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/medias"
	"tools/runtimes/db/messages"
	"tools/runtimes/db/proxys"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func DownUserVideo(c *gin.Context) {
	user, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	type reqS struct {
		Urls   []string `json:"urls"`
		UserID int64    `json:"user_id"`
	}

	var req reqS
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if req.UserID < 1 {
		response.Error(c, http.StatusBadRequest, "视频用户错误", nil)
		return
	}

	if len(req.Urls) < 1 {
		response.Error(c, http.StatusBadRequest, "未找到视频地址", nil)
		return
	}

	var proxyos []*medias.MediaUserProxy
	medias.GetDb().DB().Model(&medias.MediaUserProxy{}).Where("m_uid = ?", req.UserID).Find(&proxyos)
	pl := len(proxyos)

	var pcss []*proxy.ProxyConfig
	if pl > 0 {
		var pid int64
		if pl == 1 {
			pid = proxyos[0].ProxyID
		} else {
			pid = proxyos[rand.Intn(pl)].ProxyID
		}

		px := proxys.GetById(pid)
		if pc, err := proxy.Client(px.GetConfig(), "", px.Port, px.GetTransfer()); err == nil {
			pcss = append(pcss, pc)
		}
	}

	go func() {
		var succed int64
		var erred int64
		for _, v := range req.Urls {
			if errs := medias.GetPlatformVideos(v, pcss, "", user.Id, "", true, false); len(errs) > 0 {
				erred++
			} else {
				succed++
			}
		}
		messages.SuccessMsg(fmt.Sprintf("总下载数:%d, 成功:%d, 失败:%d", len(req.Urls), succed, erred))
	}()
	response.Success(c, nil, "")
}
