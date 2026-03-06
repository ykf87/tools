package users

import (
	"net/http"
	"strconv"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func GetInfo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	id64 := int64(id)
	mu := medias.GetMediaUserByID(id64)

	if err := mu.GetInfoFromPlatform(nil); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, mu, "")
}

func UserMeidas(c *gin.Context) {
	response.Success(c, medias.GetUserMedias(c.Query("id")), "success")
}
