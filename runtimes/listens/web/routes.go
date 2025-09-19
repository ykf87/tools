package web

import (
	"tools/runtimes/controllers/proxys"
	"tools/runtimes/controllers/ws"

	// "tools/runtimes/config"
	"tools/runtimes/controllers/user"
	// "github.com/gin-gonic/gin"
)

func router() {
	ROUTER.Static("/data", DataPath)

	ROUTER.POST("/auth/login", user.Login)

	AuthRoutes := ROUTER
	AuthRoutes.Use(AuthMiddleware)

	userGroup := AuthRoutes.Group("user")
	userGroup.GET("ws", ws.WsHandler)
	userGroup.GET("info", user.Info)
	userGroup.GET("lists", user.Lists)

	superUser := userGroup.Use(SuperAdminMiddleware)
	superUser.POST("editer", user.Editer)
	superUser.POST("delete", user.Remove)

	proxyGroup := AuthRoutes.Group("proxy")
	proxyGroup.GET("", proxys.GetList)
}
