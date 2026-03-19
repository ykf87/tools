package audios

import (
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/audios"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	var geter db.ListFinder

	if err := c.ShouldBindJSON(&geter); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	total, lists := audios.GetList(&geter)

	response.Success(c, gin.H{
		"total": total,
		"lists": lists,
	}, "")
}
