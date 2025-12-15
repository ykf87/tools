package browsers

import (
	"tools/runtimes/browser"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func Download(c *gin.Context) {
	err := browser.DownBrowserFromServer(browser.BROWSERPATH)
	response.Success(c, "", err.Error())
}
