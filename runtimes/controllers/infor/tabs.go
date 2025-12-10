package infor

import (
	"tools/runtimes/db/admins"
	"tools/runtimes/db/information"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func GetTabs(c *gin.Context) {
	var user *admins.Admin
	if u, ok := c.Get("_user"); ok {
		if us, ok := u.(*admins.Admin); ok {
			user = us
		}
	}
	response.Success(c, information.GetInforTabs(user.Id), "success")
}
