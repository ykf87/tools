package ws

import (
	"fmt"
	"net/http"
	"sync"
	"tools/runtimes/i18n"
	"tools/runtimes/ws"

	"github.com/gin-gonic/gin"
)

var CONNS sync.Map
var ConnsGroup sync.Map

type SentWsStruct struct {
	UserId  int64  `json:"user_id"`
	Group   string `json:"group"`
	Type    string `json:"type"`
	Content any    `json:"content"`
}

func init() {

}

// 连接
func Connect(c *gin.Context, id int64) (*ws.Conn, error) {
	conn, err := ws.GetConn(c.Writer, c.Request, nil, func(r *http.Request) bool {
		return true
	})

	if err != nil {
		return nil, err
	}

	CONNS.Store(id, conn)
	return conn, nil
}

// 给连接发消息
func SentMsg(id int64, date []byte) error {
	if c, ok := CONNS.Load(id); ok {
		if conn, ok := c.(*ws.Conn); ok {
			conn.WriteMessage(date)
			return nil
		}
	}
	return fmt.Errorf(i18n.T("ws: %d not Connected", id))
}

// 订阅消息
func AddGroup(name string, conn *ws.Conn) {
	if ls, ok := ConnsGroup.Load(name); ok {
		if lss, ok := ls.([]*ws.Conn); ok {
			for _, c := range lss {
				if conn.Idx == c.Idx {
					return
				}
			}
			lss = append(lss, conn)
			ConnsGroup.Store(name, lss)
		}
	}
}

// 发送订阅消息
func SentGroup(name string, data []byte) error {
	if ls, ok := ConnsGroup.Load(name); ok {
		if lss, ok := ls.([]*ws.Conn); ok {
			for _, c := range lss {
				c.WriteMessage(data)
			}
		}
	}
	return nil
}
