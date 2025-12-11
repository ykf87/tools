package infor

import (
	"tools/runtimes/db/admins"
	"tools/runtimes/db/information"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func NotRead(c *gin.Context) {
	us, ok := c.Get("_user")
	user := new(admins.Admin)
	if ok {
		user = us.(*admins.Admin)
	}

	tab := c.GetString("tab")
	page := c.GetInt("page")
	limit := c.GetInt("limit")
	rsp := information.GetNotRead(user.Id, tab, page, limit)
	response.Success(c, rsp, "success")
}
