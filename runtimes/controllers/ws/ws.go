package ws

import (
	"bytes"
	"encoding/json"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/db/admins"
	"tools/runtimes/eventbus"
	"tools/runtimes/listens/ws"

	"github.com/gin-gonic/gin"
)

var pingBytes []byte
var pongBytes []byte

func init() {
	pingBytes = []byte("ping")
	pongBytes = []byte("pong")

	eventbus.Bus.Subscribe("ws", func(data interface{}) {
		if dt, ok := data.(*ws.SentWsStruct); ok {
			obj := map[string]any{"type": dt.Type, "data": dt.Content}
			brt, err := json.Marshal(obj)
			if err == nil {
				if dt.UserId > 0 {
					ws.SentMsg(dt.UserId, brt)
				} else if dt.Group != "" {
					ws.SentGroup(dt.Group, brt)
				}
			}
		}
	})
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

	if user.Group != ""{
		for _, v := range strings.Split(user.Group, ","){
			ws.AddGroup(v, conn)
		}
		if strings.Contains(user.Group, "admin") == true{
			if config.VersionResps != nil && config.VersionResps.Code == 200 && len(config.VersionResps.Data) > 0{
				ws.SentBus(0, "version", config.VersionResps.Data, "admin")
			}
		}
	}
	eventbus.Bus.Publish("ws", map[string]any{"aaa": 1111})
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
