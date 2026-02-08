package public

import (
	"net/http"
	"strconv"
	"tools/runtimes/config"
	"tools/runtimes/db/admins"
	"tools/runtimes/i18n"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func SetWH(c *gin.Context) {
	w, _ := strconv.Atoi(c.Query("width"))
	h, _ := strconv.Atoi(c.Query("height"))

	user, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, i18n.T("Please login"), nil)
		return
	}

	config.AdminWidthAndHeight.Store(user.Id, map[string]int{
		"width":  w,
		"height": h,
	})
	response.Success(c, nil, "success")
}
