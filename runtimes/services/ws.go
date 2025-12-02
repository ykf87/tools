// 连接服务端的ws
package services

import (
	"fmt"
	"net/http"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"

	"github.com/gorilla/websocket"
)

func init() {
	dialer := websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
	}

	// 构造header
	hd := funcs.ServerHeader(config.VERSION, config.VERSIONCODE)
	header := http.Header{}
	for k, v := range hd {
		header.Set(k, v)
	}

	//发起连接
	conn, _, err := dialer.Dial(config.SERVERWS, header)
	if err != nil {
		logs.Error(err.Error())
		return
	}
	fmt.Println("ws连接成功!!!!")
	// defer conn.Close()

	// 接收消息
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			fmt.Println(string(msg))
		}
		conn.Close()
	}()
}
