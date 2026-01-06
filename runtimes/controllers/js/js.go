package js

import (
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/jses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	dt := new(db.ListFinder)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	lst, total := jses.GetJsList(dt)
	response.Success(c, gin.H{
		"list":  lst,
		"total": total,
	}, "")
}

func Tags(c *gin.Context) {
	response.Success(c, jses.GetTags(), "")
}
