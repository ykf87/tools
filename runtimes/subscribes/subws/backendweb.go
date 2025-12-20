// 客户端自己的事件通知
package subws

import (
	"fmt"
	"time"
	"tools/runtimes/browser"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/clients"
	"tools/runtimes/db/information"
	"tools/runtimes/db/proxys"
	"tools/runtimes/eventbus"
	"tools/runtimes/listens/ws"
)

func init() {
	sendws()
	message()
	closeBrowser()
	proxyChange()
	notify()
}

// 注册发送到前端的websocket事件
func sendws() {
	eventbus.Bus.Subscribe("ws", func(data any) {
		if dt, ok := data.(*ws.SentWsStruct); ok {
			obj := map[string]any{"type": dt.Type, "data": dt.Content}
			brt, err := config.Json.Marshal(obj)
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
		fmt.Println("发起message通知---")
		if dt, ok := data.([]byte); ok {
			ws.Broadcost(dt)
		}
	})
}

//	interface Notify{
//	  type: string;
//	  title: string;
//	  description?: string;
//	  content: string;
//	  meta: string;
//	  avatar?: string;
//	  closeable?: boolean;
//	  url?: string;
//	  btn?: string;
//	  method?:string;
//	}
func notify() {
	eventbus.Bus.Subscribe("notify", func(data any) {
		if dt, ok := data.([]byte); ok {
			ws.Broadcost(dt)
		}
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

// 发送消息通知给客户端
func sendInformation() {
	eventbus.Bus.Subscribe("information", func(dt any) {
		if info, ok := dt.(*information.Information); ok {
			obj := map[string]any{"type": "information", "data": info}
			if btt, err := config.Json.Marshal(obj); err == nil {
				if info.AdminId > 0 {
					ws.SentMsg(info.AdminId, btt)
				} else {
					ws.Broadcost(btt)
				}
			}
		}
	})
}
