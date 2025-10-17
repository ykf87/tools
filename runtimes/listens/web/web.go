package web

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db/admins"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/logs"
	"tools/runtimes/response"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var WebCloseCh context.CancelFunc // web关闭协程
var ROUTER *gin.Engine            // 路由
var RunPort int                   // 后端监听端口
var WebPort int                   // 网页打开的url
var WebUrl string                 // 打开的url地址
var NetIp string                  // 本机ip
var DataPath string               // 数据存储目录

type fileReplaceReqs struct {
	NewContent string
	Req        *regexp.Regexp
}

func Start(port int) {
	var err error
	NetIp, err = funcs.GetLocalIP()
	if err != nil {
		panic(err)
	}

	var ctx context.Context
	ctx, WebCloseCh = context.WithCancel(context.Background())

	gin.SetMode(gin.ReleaseMode)
	ROUTER = gin.Default()
	ROUTER.Use(Corss())
	ROUTER.Use(gzip.Gzip(gzip.DefaultCompression))
	ROUTER.Use(WriteLogs)
	ROUTER.Static("/media", config.FullPath(config.MEDIAROOT))
	RunPort = port
	router()

	// 启动服务（放到 goroutine，不阻塞主线程）
	go func() {
		ROUTER.Run(fmt.Sprintf(":%d", RunPort))
	}()
	time.Sleep(time.Second)

	// WebPort, err = funcs.FreePort()
	// if err != nil {
	// 	panic(err)
	// }
	WebPort := 15558

	go func() {
		webUrl(NetIp, RunPort, WebPort)
	}()

	WebUrl = fmt.Sprintf("http://%s:%d", NetIp, WebPort)
	DataPath = config.FullPath(config.DATAROOT)
	fmt.Println("服务已启动:", WebUrl)
	config.WebUrl = WebUrl
	config.MediaUrl = fmt.Sprintf("http://%s:%d/%s", NetIp, RunPort, config.DATAROOT)
	config.ApiUrl = fmt.Sprintf("http://%s:%d", NetIp, RunPort)

	config.MediaUrl = fmt.Sprintf("http://%s:%d/media")

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("web服务退出中...\n")
			time.Sleep(time.Second)
			return
		}
	}
}

// 允许跨域中间件
func Corss() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"*"},                                       // 允许所有来源
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // 允许所有方法
		AllowHeaders: []string{
			"*",
			// "Origin",
			// "Content-Type",
			// "Authorization",
			// "X-Requested-With",
			// "Apifoxtoken",
			// "X-Request-Id",
			// "User-Agent",
		}, // 允许所有请求头
		AllowCredentials: false,          // 不使用凭证
		MaxAge:           12 * time.Hour, // 预检请求的缓存时间
	})
}

// 日志记录
func WriteLogs(c *gin.Context) {
	logs.Logger.Info().Msgf("method: %s; path: %s", c.Request.Method, c.Request.URL.String())
	c.Next()
}

// 给web单独开一个端口
func webUrl(NetIp string, port, WebPort int) {
	webFullPath := config.FullPath(config.WEBROOT)
	assetsFullPath := filepath.Join(webFullPath, "assets")

	if _, err := os.Stat(assetsFullPath); err != nil {
		fmt.Println("Downloading web files...")
	}

	var jsFiles []string

	rps := []*fileReplaceReqs{
		&fileReplaceReqs{
			Req:        regexp.MustCompile(`VITE_SERVICE_BASE_URL:"([^"]+)"`),
			NewContent: fmt.Sprintf(`VITE_SERVICE_BASE_URL:"http://%s:%d"`, NetIp, port),
		},
		&fileReplaceReqs{
			Req:        regexp.MustCompile(`VITE_SERVICE_WS_URL:"([^"]+)"`),
			NewContent: fmt.Sprintf(`VITE_SERVICE_WS_URL:"ws://%s:%d/user/ws"`, NetIp, port),
		},
		&fileReplaceReqs{
			Req:        regexp.MustCompile(`wsurl="([^"]+)"`),
			NewContent: fmt.Sprintf(`wsurl="ws://%s:%d"`, NetIp, port),
		},
	}
	filepath.WalkDir(assetsFullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 过滤出 .js 文件
		if !d.IsDir() && filepath.Ext(path) == ".js" {
			jsFiles = append(jsFiles, path)

			if err := replaceFileContent(path, rps); err != nil {
				logs.Error(err.Error())
				panic(err)
			}
		}
		return nil
	})
	if len(jsFiles) < 1 {
		logs.Error("No web files!")
		panic("No web files!")
	}

	r := gin.Default()
	r.Static("/", config.FullPath(config.WEBROOT))
	r.NoRoute(func(c *gin.Context) {
		c.File(config.FullPath(config.WEBROOT, "index.html"))
	})

	r.Run(fmt.Sprintf(":%d", WebPort))
}

// 替换文件内容
func replaceFileContent(file string, regs []*fileReplaceReqs) error {
	// 读取文件内容
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	newContent := string(content)
	for _, v := range regs {
		newContent = v.Req.ReplaceAllString(newContent, v.NewContent)
	}
	// newContent := re.ReplaceAllString(string(content), newcont)

	err = os.WriteFile(file, []byte(newContent), 0644)
	if err != nil {
		return err
	}
	return nil
}

// 判断是否来自ws
func isWebSocket(c *gin.Context) bool {
	connectionHeader := strings.ToLower(c.GetHeader("Connection"))
	upgradeHeader := strings.ToLower(c.GetHeader("Upgrade"))

	return strings.Contains(connectionHeader, "upgrade") && upgradeHeader == "websocket"
}

// 用户鉴权
func AuthMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" && isWebSocket(c) {
		token = c.Query("token")
	}
	if token == "" {
		response.Error(c, http.StatusUnauthorized, i18n.T("Please Login first"), nil)
		c.Abort()
		return
	}

	adm, err := admins.GetAdminFromJwt(token)
	if err != nil {
		logs.Error(err.Error())
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
		c.Abort()
		return
	}

	c.Set("_user", adm)
	c.Next()
}

// 超级用户权限中间件
func SuperAdminMiddleware(c *gin.Context) {
	u, ok := c.Get("_user")
	if !ok {
		response.Error(c, http.StatusBadRequest, i18n.T("Please Login first"), nil)
		c.Abort()
		return
	}
	us, ok := u.(*admins.Admin)
	if !ok || us.Id < 1 || us.Main != 1 {
		response.Error(c, http.StatusBadRequest, i18n.T("You do not have permission to perform this operation"), nil)
		c.Abort()
		return
	}
	c.Next()
}
