// 客户端自己的事件通知
package subscribes

import (
	"time"
	"tools/runtimes/browser"
	"tools/runtimes/db"
	"tools/runtimes/db/clients"
	"tools/runtimes/db/proxys"
	"tools/runtimes/eventbus"
	"tools/runtimes/listens/ws"

	jsoniter "github.com/json-iterator/go"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	sendws()
	message()
	closeBrowser()
	proxyChange()
}

// 注册发送到前端的websocket事件
func sendws() {
	eventbus.Bus.Subscribe("ws", func(data any) {
		if dt, ok := data.(*ws.SentWsStruct); ok {
			obj := map[string]any{"type": dt.Type, "data": dt.Content}
			brt, err := Json.Marshal(obj)
			if err == nil {
				if dt.UserId > 0 {
					ws.SentMsg(dt.UserId, brt)
				} else if dt.Group != "" {
					ws.SentGroup(dt.Group, brt)
				} else {
					ws.Broadcost(brt)
				}
			}
		}
	})
}

// 注册前端message通知的中间件,并且发送message类型的ws消息给前端
func message() {
	eventbus.Bus.Subscribe("message", func(data any) {

	})
}

// 浏览器被关闭的事件
func closeBrowser() {
	eventbus.Bus.Subscribe("browser-close", func(dt any) {
		if bu, ok := dt.(*browser.User); ok {
			if bu.Id > 0 {
				browser.Running.Delete(bu.Id)
				if bs, err := clients.GetBrowserById(bu.Id); err == nil {
					msgdata := new(ws.SentWsStruct)
					msgdata.UserId = bu.UserId
					msgdata.Type = "browser"
					msgdata.Content = bs
					eventbus.Bus.Publish("ws", msgdata)
				}
			}
		}
	})
}

// 监听代理改变事件,同步修改浏览器的local
func proxyChange() {
	eventbus.Bus.Subscribe("proxy_change", func(dt any) {
		if proxy, ok := dt.(*proxys.Proxy); ok {
			go func() {
				time.Sleep(time.Second * 1)
				db.DB.Model(&clients.Browser{}).Where("proxy = ?", proxy.Id).Updates(map[string]any{
					"local":      proxy.Local,
					"ip":         proxy.Ip,
					"lang":       proxy.Lang,
					"timezone":   proxy.Timezone,
					"proxy_name": proxy.Name,
				})
			}()
		}
	})
}
