package browsers

import (
	"tools/runtimes/bs"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func Download(c *gin.Context) {
	err := bs.DownBrowserBinFile(bs.BROWSERPATH)
	response.Success(c, "", err.Error())
}
