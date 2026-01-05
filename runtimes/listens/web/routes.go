package web

import (
	"tools/runtimes/controllers/browsers"
	"tools/runtimes/controllers/infor"
	"tools/runtimes/controllers/phones"
	"tools/runtimes/controllers/proxys"
	"tools/runtimes/controllers/suggs"
	"tools/runtimes/controllers/tags"
	"tools/runtimes/controllers/task"
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
	ROUTER.POST("sugg_cate", suggs.SuggCate)
	ROUTER.POST("browser/download", browsers.Download)
	ROUTER.GET("client/ws", phones.Ws)   // app 的ws连接
	ROUTER.GET("client/api", phones.Api) // app 的api连接
	AuthRoutes := ROUTER
	AuthRoutes.Use(AuthMiddleware)

	userGroup := AuthRoutes.Group("user")
	{
		userGroup.GET("ws", ws.WsHandler)
		userGroup.GET("info", user.Info)
		userGroup.GET("lists", user.Lists)
		userGroup.POST("repwd", user.RePassword)
		userGroup.POST("suggestion", user.Suggestion)
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

	clientGroup := AuthRoutes.Group("phones")
	{
		clientGroup.POST("", phones.List)
		clientGroup.GET("total", phones.Total)
		clientGroup.GET("urls", phones.ConnUrl)
		clientGroup.GET("tags", phones.PhoneTags)
		clientGroup.POST("add", phones.Editer)
		clientGroup.POST("/:id", phones.Editer)
		clientGroup.POST("delete/:id", phones.Delete)
		// clientGroup.GET("api", phones.Api)
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

	suggGroup := AuthRoutes.Group("sugg")
	{
		suggGroup.POST("add", suggs.AddSuggestion)
	}

	// 消息通知
	inforGroup := AuthRoutes.Group("infor")
	{
		inforGroup.GET("tabs", infor.GetTabs)
		inforGroup.GET("notread", infor.NotRead)
	}

	// 任务
	taskGroup := AuthRoutes.Group("tasks")
	{
		taskGroup.POST("", task.List)
		taskGroup.GET("viewbasedata", task.BaseData)
		taskGroup.POST("ae", task.AddOrEdit)
		taskGroup.GET("devices", task.TaskDevices)
	}
}
