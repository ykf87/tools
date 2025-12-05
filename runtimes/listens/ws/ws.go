package ws

import (
	"fmt"
	"net/http"
	"sync"
	"tools/runtimes/eventbus"
	"tools/runtimes/i18n"
	"tools/runtimes/ws"

	"github.com/gin-gonic/gin"
)

var CONNS sync.Map
var ConnsGroup sync.Map

type SentWsStruct struct {
	UserId  int64  `json:"user_id"` // 给某个用户发送消息
	Group   string `json:"group"`   // 给某个分组发送消息
	Type    string `json:"type"`    // 消息类型
	Content any    `json:"content"` // 消息主体
}

func init() {

}

// 组装并发布ws事件
// uid 当前登录的用户id
// tp 消息类型, 须于前端接收的对应
// content 消息体, 前端解析
// group 接收消息的分组
func SentBus(uid int64, tp string, content any, group string) {
	dt := new(SentWsStruct)
	dt.Type = tp
	dt.UserId = uid
	dt.Content = content
	dt.Group = group

	dt.Send()
}

// 发布ws
func (this *SentWsStruct) Send() {
	eventbus.Bus.Publish("ws", this)
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
	} else {
		newls := []*ws.Conn{conn}
		ConnsGroup.Store(name, newls)
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

// 广播消息
func Broadcost(data []byte) {
	CONNS.Range(func(k, v any) bool {
		if conn, ok := v.(*ws.Conn); ok {
			conn.WriteMessage(data)
		}
		return true
	})
}
