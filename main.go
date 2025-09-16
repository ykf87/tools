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
 **/

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	_ "tools/runtimes"
	_ "tools/runtimes/db"
	"tools/runtimes/listens/web"
)

func main() {
	// checkNode()

	// port, err := funcs.FreePort()
	// if err != nil {
	// 	panic(err)
	// }
	port := 19998
	go web.Start(port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("")

	web.WebCloseCh()

	fmt.Println("----")
}
