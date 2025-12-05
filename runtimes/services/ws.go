// 连接服务端的ws
package services

import (
	"fmt"
	"net/http"
	"tools/runtimes/config"
	"tools/runtimes/eventbus"
	"tools/runtimes/funcs"
	"tools/runtimes/ws"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type Client struct {
	Url     string
	Proxy   string
	Headers map[string]string
	Conn    *websocket.Conn
}

var WSClient *ws.Client

func init() {
	hd := http.Header{}
	for k, v := range funcs.ServerHeader(config.VERSION, config.VERSIONCODE) {
		hd.Set(k, v)
	}

	WSClient = ws.New(config.SERVERWS, nil, ws.EventHandler{
		OnOpen: func() {
			// fmt.Println("服务连接成功")
			// WSClient.Send([]byte("ping"))
		},
		OnError: func(err error) {
			// logs.Error("服务端ws错误:" + err.Error())
		},
		OnClose:   func() { fmt.Println("关闭连接") },
		OnMessage: readMessage,
		ConnHead: func() http.Header {
			hd := http.Header{}
			for k, v := range funcs.ServerHeader(config.VERSION, config.VERSIONCODE) {
				hd.Set(k, v)
			}
			return hd
		},
	})
	WSClient.Start()

	// 以下是测试代码-----------
	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 5)
	// 		fmt.Println("发送消息")
	// 		WSClient.Send([]byte(`{"tp":"test","data":"sdfsdfkfsdkf"}`))
	// 	}
	// }()
}

// 读取消息
func readMessage(messageType int, message []byte) {
	if messageType != websocket.PingMessage {
		msg := gjson.ParseBytes(message)
		eventbus.Bus.Publish(msg.Get("tp").String(), msg.Get("data").String())
	}
}
