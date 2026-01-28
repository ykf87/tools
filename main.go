package main

/**
# * Windows (amd64)
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
go build -ldflags "-s -w" -o tools.exe
go build -ldflags "-s -w" -o tools.exe
减少某些安全软件误报的可能。
**/

import (
	"fmt"
	"strings"
	"time"
	_ "tools/runtimes"
	"tools/runtimes/config"
	_ "tools/runtimes/db"
	"tools/runtimes/db/tasks"
	"tools/runtimes/i18n"
	"tools/runtimes/listens/web"
	"tools/runtimes/mainsignal"
	_ "tools/runtimes/minio"
	_ "tools/runtimes/subscribes/submqs"
	_ "tools/runtimes/subscribes/subws"
	"tools/runtimes/syncuuid"
	// "tools/runtimes/mq"
)

func main() {
	if checkLocal() != nil {
		time.Sleep(time.Second * 10)
		return
	}

	port := 19998

	fmt.Println("欢迎使用小卡卡辅助工具.有任何问题可以随时在系统内联系开发者或者前往官网留言.祝您使用愉快!")
	fmt.Println("系统UUID:", syncuuid.MachineUUID())
	go web.Start(port)
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit
	<-mainsignal.MainCtx.Done()
	mainsignal.MainStop()
	fmt.Println("\n系统准备关闭,释放内存中,请稍后...")
	flush()

	fmt.Println("系统已退出,感谢您的使用!")
}

func checkLocal() error {
	if strings.HasPrefix(strings.ToLower(config.RuningRoot), "c:") {
		msg := i18n.T("禁止在C盘启动")
		fmt.Println(msg)
		fmt.Println("由于软件使用过程中会产生许多数据,而C盘作为系统盘并不建议存储业务数据,因此我们限制将软件放在C盘使用.")
		time.Sleep(time.Second * 30)
		return fmt.Errorf("error")
	}
	return nil
}

func flush() {
	tasks.Seched.Stop()
	// bs.Flush()
	web.WebCloseCh()
	time.Sleep(time.Second)
}
