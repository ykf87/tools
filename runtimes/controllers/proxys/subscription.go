package proxys

import (
	"fmt"
	"net/http"
	"tools/runtimes/config"
	"tools/runtimes/response"
	"tools/runtimes/services"

	"github.com/gin-gonic/gin"
)

func Subscription(c *gin.Context){
	suburl := fmt.Sprint(config.SERVERDOMAIN, "subscription")

	proxys, err := services.GerProxySub(suburl)
	if err != nil{
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, proxys, "success")
}
