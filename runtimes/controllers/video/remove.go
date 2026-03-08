package video

import (
	"net/http"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func RemoveVideo(c *gin.Context) {
	type idss struct {
		Ids []int64 `json:"ids"`
	}

	ids := new(idss)
	if err := c.ShouldBindJSON(ids); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := medias.DeleteMediaFiles(ids.Ids); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, nil, "")
}
