// 连接服务端的ws
package services

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"

	"github.com/gorilla/websocket"
)

type Client struct {
	Url     string
	Proxy   string
	Headers map[string]string
	Conn    *websocket.Conn
}

func init() {
	// dialer := websocket.Dialer{
	// 	Proxy: http.ProxyFromEnvironment,
	// }

	// // 构造header
	// hd := funcs.ServerHeader(config.VERSION, config.VERSIONCODE)
	// header := http.Header{}
	// for k, v := range hd {
	// 	header.Set(k, v)
	// }

	// //发起连接
	// conn, _, err := dialer.Dial(config.SERVERWS, header)
	// if err != nil {
	// 	logs.Error(err.Error())
	// 	return
	// }
	// fmt.Println("ws连接成功!!!!")
	// // defer conn.Close()

	// // 接收消息
	// go func() {
	// 	defer conn.Close()
	// 	for {
	// 		_, msg, err := conn.ReadMessage()
	// 		if err != nil {
	// 			break
	// 		}
	// 		fmt.Println(string(msg))
	// 	}
	// }()
	cli := &Client{
		Url:     config.SERVERWS,
		Headers: funcs.ServerHeader(config.VERSION, config.VERSIONCODE),
	}
	if err := cli.Connect(); err != nil {
		logs.Error(err.Error())
	}
}

// 发起连接
func (this *Client) Connect() error {
	if this.Url == "" {
		return errors.New("url不能为空!")
	}
	var proxy *url.URL
	if this.Proxy != "" {
		if proxyURL, err := url.Parse(this.Proxy); err == nil {
			proxy = proxyURL
		}
	}
	dialer := websocket.Dialer{
		Proxy:            http.ProxyURL(proxy),
		HandshakeTimeout: 10 * time.Second,
	}

	header := http.Header{}
	for k, v := range this.Headers {
		header.Set(k, v)
	}

	conn, _, err := dialer.Dial(this.Url, header)
	if err != nil {
		return err
	}
	this.Conn = conn

	go this.listen()
	return nil
}

// 监听ws并且断线重连、心跳等
func (this *Client) listen() {
	fmt.Println("监听---")
}
