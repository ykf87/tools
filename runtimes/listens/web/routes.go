package web

import (
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
	userGroup.Use(AuthMiddleware)
	userGroup.GET("info", user.Info)
	userGroup.GET("lists", user.Lists)
}
