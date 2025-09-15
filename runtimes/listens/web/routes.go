package web

import (
	"fmt"
	// "tools/runtimes/config"
	"tools/runtimes/controllers/user"
	// "github.com/gin-gonic/gin"
)

func router() {
	fmt.Println(DataPath, "----")
	ROUTER.Static("/data", DataPath)
	// ROUTER.Static("/web", config.FullPath(".web"))
	// // 捕获未匹配的路由 → index.html
	// ROUTER.NoRoute(func(c *gin.Context) {
	// 	c.File(config.FullPath(".web", "index.html"))
	// })

	ROUTER.POST("/login", user.Login)
	// ROUTER.Group("user")
}
