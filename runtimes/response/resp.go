package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data any, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": data,
		"msg":  msg,
	})
	return
}

func Error(c *gin.Context, code int, msg string, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"data": data,
		"msg":  msg,
	})
	return
}
