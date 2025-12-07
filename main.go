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
	"os"
	"os/signal"
	"syscall"
	_ "tools/runtimes"
	"tools/runtimes/browser"
	_ "tools/runtimes/db"
	"tools/runtimes/listens/web"
	_ "tools/runtimes/listens/ws/subscribes"
)

func main() {
	port := 19998
	go web.Start(port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("")

	web.WebCloseCh()
	browser.Flush()

	fmt.Println("----")
}
