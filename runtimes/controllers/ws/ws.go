package ws

import (
	"bytes"
	"fmt"
	"net/http"
	"tools/runtimes/config"
	"tools/runtimes/db/admins"
	"tools/runtimes/eventbus"
	cws "tools/runtimes/ws"

	"github.com/gin-gonic/gin"
)

var pingBytes []byte
var pongBytes []byte

func init() {
	pingBytes = []byte("ping")
	pongBytes = []byte("pong")

	eventbus.Bus.Subscribe("ws", func(data interface{}) {
		cws.SendContent("1", "download", data)
	})
}

func WsHandler(c *gin.Context) {
	conn, err := cws.GetConn(c.Writer, c.Request, nil, func(r *http.Request) bool {
		return true
	})

	if err != nil {
		return
	}

	u, ok := c.Get("_user")
	if !ok {
		return
	}
	user := u.(*admins.Admin)
	cws.CONNS.Store(fmt.Sprintf("%d", user.Id), conn)

	conn.WriteMessage([]byte(fmt.Sprintf(`{"type":"version", "data":"%s"}`, config.VERSION)))
	eventbus.Bus.Publish("ws", map[string]any{"aaa": 1111})
	for {
		p, err := conn.ReadMessage()
		// messageType, p, err := conn.Conn.ReadMessage()
		if err != nil {
			break
		}

		if bytes.Equal(p, pingBytes) {
			conn.WriteMessage(pongBytes)
			// conn.Conn.WriteMessage(messageType, pongBytes)
			continue
		}
	}
}
