package web

import (
	"tools/runtimes/controllers/proxys"
	"tools/runtimes/controllers/tags"
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
	{
		userGroup.GET("ws", ws.WsHandler)
		userGroup.GET("info", user.Info)
		userGroup.GET("lists", user.Lists)
	}

	superUser := userGroup.Use(SuperAdminMiddleware)
	{
		superUser.POST("editer", user.Editer)
		superUser.POST("delete", user.Remove)
	}

	proxyGroup := AuthRoutes.Group("proxy")
	{
		proxyGroup.GET("", proxys.GetList)
		proxyGroup.POST("", proxys.Editer)
		proxyGroup.POST("/:id", proxys.Editer)
		proxyGroup.GET("/:id", proxys.GetRow)
		proxyGroup.POST("local/", proxys.GetLocal)
		proxyGroup.POST("local/:id", proxys.GetLocal)
		proxyGroup.POST("delete/:id", proxys.Remove)
	}

	tagGroup := AuthRoutes.Group("tags")
	{
		tagGroup.GET("", tags.List)
		tagGroup.POST("add", tags.Add)
		tagGroup.POST("delete/:id", tags.Delete)
	}
}
