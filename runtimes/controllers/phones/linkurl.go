package phones

import (
	"fmt"
	"net/http"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// 获取手机端连接的地址
func ConnUrl(c *gin.Context) {
	ip, err := funcs.GetLocalIP(true)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	urls := map[string]string{
		"ws":  fmt.Sprintf("ws://%s:%d/client/ws", ip, config.ApiPort),
		"api": fmt.Sprintf("http://%s:%d/client/api", ip, config.ApiPort),
		"app": fmt.Sprintf("%sdown/app@last", config.SERVERDOMAIN),
	}
	if wsqr, err := qrcode.Encode(urls["ws"], qrcode.Medium, 256); err == nil {
		urls["wsqr"] = funcs.Base64Encode(string(wsqr))
	}
	if wsqr, err := qrcode.Encode(urls["api"], qrcode.Medium, 256); err == nil {
		urls["apiqr"] = funcs.Base64Encode(string(wsqr))
	}
	if wsqr, err := qrcode.Encode(urls["app"], qrcode.Medium, 256); err == nil {
		urls["appqr"] = funcs.Base64Encode(string(wsqr))
	}
	response.Success(c, urls, "success")
}
