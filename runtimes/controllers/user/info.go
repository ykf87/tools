package user

import (
	"net/http"
	"tools/runtimes/db/admins"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func Info(c *gin.Context) {
	u, ok := c.Get("_user")
	if !ok {
		response.Error(c, http.StatusNotFound, i18n.T("Please Login first"), nil)
		return
	}
	adm := u.(*admins.Admin)
	r, _ := parses.Marshal(adm, c)
	response.Success(c, r, "success")
}
