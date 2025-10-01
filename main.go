package main

/**
# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o

# Windows (arm64) - 如果需要支持 Win11 on ARM
GOOS=windows GOARCH=arm64 go build -o

# macOS Intel (amd64)
GOOS=darwin GOARCH=amd64 go build -o

# macOS Apple Silicon (M1/M2, arm64)
GOOS=darwin GOARCH=arm64 go build -o

# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o

# Linux (arm64) - 树莓派/服务器常见
GOOS=linux GOARCH=arm64 go build -o

# 32位操作系统
go build -o tools.exe -ldflags "-s -w" main.go


1️⃣ 强制使用旧 PE 兼容
在编译时加 GOAMD64=v1：
set GOAMD64=v1
go build -o tools.exe main.go
GOAMD64=v1 会生成兼容 较老 CPU/OS 的二进制，避免新 PE 头特性。
对于 Windows 10 64 位，通常完全可以运行。

2️⃣ 使用 -ldflags "-s -w" 减小文件并关闭调试信息
go build -ldflags "-s -w" -o tools.exe main.go
go build -ldflags "-s -w" -o tools.exe main.go
减少某些安全软件误报的可能。
 **/

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	// "time"
	_ "tools/runtimes"
	_ "tools/runtimes/db"

	// "tools/runtimes/db/mqs"
	"tools/runtimes/listens/web"
	// "tools/runtimes/mq"
	// "tools/runtimes/proxy"
)

func main() {
	// prostr := "ss://YWVzLTEyOC1nY206NTU0MTkyOTgtNjFiMC00OWVlLTkyZjgtOTM5MDQ1ZDY1N2Mz@8aqaqmlb.sched.sma-2.1.ssndlls.xin:40011#%F0%9F%87%B9%F0%9F%87%BC%E9%AB%98%E7%BA%A7%20%7C%20%E5%8F%B0%E6%B9%BE%2001"
	// prostr = "ss://YWVzLTEyOC1nY206NTU0MTkyOTgtNjFiMC00OWVlLTkyZjgtOTM5MDQ1ZDY1N2Mz@4dqaqmlb.sched.sma-1.2.ssndlls.xin:30061#%F0%9F%87%B2%F0%9F%87%BD%E6%A0%87%E5%87%86%20%7C%20%E5%A2%A8%E8%A5%BF%E5%93%A5"
	// prostr = "socks://127.0.0.1:33211"
	// pp, err := proxy.Run(prostr, "127.0.0.1", 15586, false)
	// fmt.Println(err, pp.Listened())
	// fmt.Println(err, "-------------------run err")
	// checkNode()

	// port, err := funcs.FreePort()
	// if err != nil {
	// 	panic(err)
	// }

	// mqSystem := mq.New(mqs.NewStore(db.MQDB))
	// mqSystem.Subscribe("download", func(msg *mqs.Mq) error {
	// 	fmt.Printf("processing message %d: %s", msg.ID, msg.Payload)
	// 	// 模拟下载
	// 	time.Sleep(2 * time.Second)
	// 	if msg.RetryCount < 1 {
	// 		return fmt.Errorf("simulate fail once")
	// 	}
	// 	return nil
	// })
	// // mqSystem.Start()
	// mqSystem.Publish("download", "http://example.com/file1.zip", 0)

	// vbPath := browser.BROWSERFILE

	// // 指定自定义浏览器路径
	// url := launcher.New().
	// 	Leakless(false).
	// 	Bin(vbPath).     // 使用 VirtualBrowser
	// 	Headless(false). // 是否显示窗口
	// 	Set("user-data-dir", config.BROWSERCACHE+"/1").
	// 	Set("worker-id", "1").
	// 	MustLaunch()

	// browser := rod.New().ControlURL(url).MustConnect()
	// defer browser.MustClose()

	// page := browser.MustPage("https://www.baidu.com")
	// title := page.MustEval("() => document.title").String()
	// fmt.Println("Page title:", title)

	port := 19998
	go web.Start(port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("")

	web.WebCloseCh()

	fmt.Println("----")
}
