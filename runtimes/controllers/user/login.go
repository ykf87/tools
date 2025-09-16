package user

import (
	"net/http"
	"tools/runtimes/db/admins"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type LoginData struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	// account := c.PostForm("userName")
	// password := c.PostForm("password")
	ld := new(LoginData)
	if err := c.ShouldBind(ld); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	adm, err := admins.Login(ld.UserName, ld.Password)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	response.Success(c, map[string]any{
		"token":        adm.Jwt,
		"refreshToken": adm.Jwt,
	}, "success")
}
