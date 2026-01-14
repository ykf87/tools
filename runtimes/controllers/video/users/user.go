package users

import (
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	var user *admins.Admin
	if u, ok := c.Get("_user"); ok {
		if user, ok = u.(*admins.Admin); !ok {
			response.Error(c, http.StatusBadRequest, "请登录", nil)
			return
		}
	}
	if user == nil || user.Id < 1 {
		response.Error(c, http.StatusBadRequest, "请登录", nil)
		return
	}

	dt := new(db.ListFinder)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	list, total := medias.GetMediaUsers(user.Id, dt)
	rsp := gin.H{
		"list":  list,
		"total": total,
	}

	response.Success(c, rsp, "")
}

func GetTags(c *gin.Context) {
	response.Success(c, medias.GetTags(), "")
}
func GetPlatforms(c *gin.Context) {
	response.Success(c, medias.GetUserPlatforms(), "")
}

func Delete(c *gin.Context) {

}
