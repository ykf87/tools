package proxys

import (
	"fmt"
	"net/http"
	"sync"
	"tools/runtimes/db"
	"tools/runtimes/db/proxys"
	"tools/runtimes/i18n"
	"tools/runtimes/proxy"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type GetLocalData struct {
	Config string `json:"config" form:"config"`
	Trans  string `json:"trans" form:"trans"`
}

func GetLocal(c *gin.Context) {
	var config string
	var trans string
	id := c.Param("id")
	if id != "" {
		px := proxys.GetById(id)
		if px != nil && px.Id > 0 {
			config = px.Config
			trans = px.Transfer
		}
	} else {
		ddt := new(GetLocalData)
		if err := c.ShouldBind(ddt); err != nil {
			response.Error(c, http.StatusBadRequest, "", nil)
			return
		}
		config = ddt.Config
		trans = ddt.Trans
	}

	if config == "" {
		response.Error(c, http.StatusBadRequest, i18n.T("Request data is empty"), nil)
		return
	}

	local, err := proxy.GetLocal(config, trans)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, gin.H{
		"local":    local,
		"localico": fmt.Sprintf("https://flagpedia.net/data/flags/h80/%s.webp", local),
	}, "")
}

// 获取地区
func Local(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, i18n.T("Error"), nil)
		return
	}
	pc := proxys.GetById(id)
	if pc != nil && pc.Id > 0 {
		loc, err := proxy.GetLocal(pc.Config, pc.Transfer)
		if err == nil {
			pc.Local = loc
			pc.Save(nil)
		} else {
			response.Error(c, http.StatusBadRequest, i18n.T("Region acquisition failed"), nil)
			return
		}
	}
	response.Success(c, nil, "")
	return
}

// 批量获取地区
func Locals(c *gin.Context) {
	ids := new(rmvsdt)
	if err := c.ShouldBind(ids); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	if len(ids.Ids) > 0 {
		var px []*proxys.Proxy
		db.DB.Model(&proxys.Proxy{}).Where("id in ?", ids.Ids).Find(&px)
		if len(px) > 0 {
			var wg sync.WaitGroup
			for _, v := range px {
				wg.Add(1)
				go func() {
					defer wg.Done()
					proxy.GetLocal(v.Config, v.Transfer)
				}()
			}

			wg.Wait()
		}
	}
	response.Success(c, nil, "")
}
