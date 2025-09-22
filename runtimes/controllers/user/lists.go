package user

import (
	"strconv"
	"tools/runtimes/db/admins"

	// "tools/runtimes/downloader"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func Lists(c *gin.Context) {
	// go downloader.Down("https://nbg1-speed.hetzner.com/100MB.bin", "./aaa.bin", "")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	q := c.DefaultQuery("q", "")
	bykey := c.DefaultQuery("k", "")
	by := c.DefaultQuery("by", "")

	list, total := admins.AdminList(page, limit, q, bykey, by)

	dt, _ := parses.Marshal(gin.H{
		"list":  list,
		"total": total,
	}, c)
	response.Success(c, dt, "success")
}
