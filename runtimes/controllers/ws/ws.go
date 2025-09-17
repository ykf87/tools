package ws

import (
	"bytes"
	"fmt"
	"net/http"
	"tools/runtimes/config"
	cws "tools/runtimes/ws"

	"github.com/gin-gonic/gin"
)

var pingBytes []byte
var pongBytes []byte

func init() {
	pingBytes = []byte("ping")
	pongBytes = []byte("pong")
}

func WsHandler(c *gin.Context) {
	conn, err := cws.GetConn(c.Writer, c.Request, nil, func(r *http.Request) bool {
		return true
	})

	if err != nil {
		return
	}

	conn.WriteMessage([]byte(fmt.Sprintf(`{"type":"version", "data":"%s"}`, config.VERSION)))
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
