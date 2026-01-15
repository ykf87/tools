// 客户端自己的事件通知
package subws

import (
	"fmt"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/information"
	"tools/runtimes/db/notifies"
	"tools/runtimes/db/proxys"
	"tools/runtimes/eventbus"
	"tools/runtimes/listens/ws"
	"tools/runtimes/logs"
)

func init() {
	sendws()
	message()
	closeBrowser()
	proxyChange()
	notify()
	proxyPing()
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

// 消息通知
func notify() {
	eventbus.Bus.Subscribe("notify", func(data any) {
		if dt, ok := data.(*notifies.Notify); ok {
			if dt.Type == "" {
				dt.Type = "info"
			}
			if dt.Meta == "" {
				dt.Meta = time.Now().Format("2006-01-02 15:04:05")
			}
			byteData, err := config.Json.Marshal(map[string]any{
				"type": "notify",
				"data": dt,
			})
			if err != nil {
				logs.Error("notify data parse error:" + err.Error())
				return
			}
			if dt.AdminID > 0 {
				ws.SentMsg(dt.AdminID, byteData)
			} else {
				ws.Broadcost(byteData)
			}
		}
	})
}

// 浏览器被关闭的事件
func closeBrowser() {
	eventbus.Bus.Subscribe("browser-close", func(dt any) {
		if bu, ok := dt.(*browserdb.Browser); ok {
			if bu.Id > 0 {
				// browser.Running.Delete(bu.Id)
				if bs, err := browserdb.GetBrowserById(bu.Id); err == nil {
					msgdata := new(ws.SentWsStruct)
					msgdata.UserId = bu.AdminID
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
				db.DB.Model(&browserdb.Browser{}).Where("proxy = ?", proxy.Id).Updates(map[string]any{
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

// 代理ping结果
func proxyPing() {
	eventbus.Bus.Subscribe("proxy-ping", func(dt any) {
		if info, ok := dt.(proxys.PingResp); ok {
			mmv := map[string]any{
				"type": "proxy-ping",
				"data": info.Ping,
			}
			if msg, err := config.Json.Marshal(mmv); err == nil {
				ws.SentMsg(info.UID, msg)
			}
		}
	})
}
