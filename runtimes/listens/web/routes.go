package web

import (
	"tools/runtimes/controllers/browsers"
	"tools/runtimes/controllers/proxys"
	"tools/runtimes/controllers/tags"
	"tools/runtimes/controllers/user"
	"tools/runtimes/controllers/video/down"
	"tools/runtimes/controllers/ws"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func router() {
	ROUTER.GET("", func(c *gin.Context) {
		response.Success(c, nil, "")
	})
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
		proxyGroup.POST("batchadd", proxys.BatchAdd)
		proxyGroup.POST("batchediter", proxys.BatchEditer)
		proxyGroup.GET("/:id", proxys.GetRow)
		proxyGroup.POST("locals", proxys.Locals)
		proxyGroup.POST("local/:id", proxys.Local)
		proxyGroup.POST("delete", proxys.Removes)
		proxyGroup.POST("delete/:id", proxys.Remove)
		proxyGroup.POST("start/:id", proxys.Start)
		proxyGroup.POST("stop/:id", proxys.Stop)
		proxyGroup.POST("ping/:id", proxys.Ping)
		proxyGroup.POST("subscription", proxys.Subscription)
		// proxyGroup.POST("local/:id", proxys.Local)
	}

	tagGroup := AuthRoutes.Group("tags")
	{
		tagGroup.GET("", tags.List)
		tagGroup.POST("add", tags.Add)
		tagGroup.POST("delete/:id", tags.Delete)
	}

	videosGroup := AuthRoutes.Group("video")
	{
		vDownloader := videosGroup.Group("downloader")
		{
			vDownloader.POST("", down.Download)
		}
	}

	browserGroup := AuthRoutes.Group("browser")
	{
		browserGroup.GET("", browsers.List)
		browserGroup.GET("langs", browsers.GetLangs)
		browserGroup.GET("timezones", browsers.GetTimezones)
		browserGroup.GET("tags", browsers.BrowserTags)
		browserGroup.POST("", browsers.Editer)
		browserGroup.POST("/:id", browsers.Editer)
		browserGroup.POST("start/:id", browsers.Start)
		browserGroup.POST("stop/:id", browsers.Stop)
		browserGroup.POST("delete/:id", browsers.Delete)
	}

	spiderGroup := AuthRoutes.Group("spider")
	{
		spiderVideoGroup := spiderGroup.Group("video")
		{
			spiderVideoGroup.POST("", down.List)
			spiderVideoGroup.POST("down", down.Download)
			spiderVideoGroup.POST("mkdir", down.Mkdir)
			spiderVideoGroup.POST("open", down.OpenDir)
		}
	}
}
