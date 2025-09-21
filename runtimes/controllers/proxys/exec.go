package proxys

import (
	"fmt"
	"net/http"
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
