package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
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
			if brt, err := json.Marshal(obj); err == nil {
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

	ws.SentMsg(user.Id, []byte(fmt.Sprintf(`{"type":"version", "data":"%s"}`, config.VERSION)))
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
