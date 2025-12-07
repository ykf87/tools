package ws

import (
	"bytes"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/db/admins"
	"tools/runtimes/listens/ws"
	"tools/runtimes/services"

	"github.com/gin-gonic/gin"
)

var pingBytes []byte
var pongBytes []byte

func init() {
	pingBytes = []byte("ping")
	pongBytes = []byte("pong")

	// eventbus.Bus.Subscribe("ws", func(data any) {
	// 	if dt, ok := data.(*ws.SentWsStruct); ok {
	// 		obj := map[string]any{"type": dt.Type, "data": dt.Content}
	// 		brt, err := json.Marshal(obj)
	// 		if err == nil {
	// 			if dt.UserId > 0 {
	// 				ws.SentMsg(dt.UserId, brt)
	// 			} else if dt.Group != "" {
	// 				ws.SentGroup(dt.Group, brt)
	// 			} else {
	// 				ws.Broadcost(brt)
	// 			}
	// 		}
	// 	}
	// })
}

func WsHandler(c *gin.Context) {
	u, ok := c.Get("_user")
	if !ok {
		return
	}
	user := u.(*admins.Admin)

	conn, err := ws.Connect(c, user.Id)
	if err != nil {
		return
	}

	// ws.SentMsg(user.Id, []byte(fmt.Sprintf(`{"type":"version", "data":"%s"}`, config.VERSION)))

	if user.Group != "" {
		for _, v := range strings.Split(user.Group, ",") {
			ws.AddGroup(v, conn)
		}
		if strings.Contains(user.Group, "admin") == true {
			if services.VersionResps != nil && services.VersionResps.Code == 200 && len(services.VersionResps.Data) > 0 {
				rsps := gin.H{
					"code":        config.VERSION,
					"code_number": config.VERSIONCODE,
					"versions":    services.VersionResps.Data,
				}
				for _, v := range services.VersionResps.Data {
					if v.CodeNum == config.VERSIONCODE {
						rsps["title"] = v.Title
						rsps["released"] = v.Released
						rsps["desc"] = v.Desc
						rsps["content"] = v.Content
						rsps["released_time"] = v.ReleaseTime
						break
					}
				}
				ws.SentBus(0, "version", rsps, "admin")
			}
		}
	}
	// eventbus.Bus.Publish("ws", map[string]any{"aaa": 1111})
	for {
		p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if bytes.Equal(p, pingBytes) {
			conn.WriteMessage(pongBytes)
			continue
		}
	}
}
