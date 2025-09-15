package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
	"tools/runtimes/i18n"

	"github.com/fasthttp/websocket"
)

var CONNS sync.Map
var Idx int64

type Conn struct {
	Conn    *websocket.Conn
	Idx     int64
	Addtime int64
}

type WsResp struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Notify struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type Message struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Plain   bool   `json:"plain"`
}

func GetConn(w http.ResponseWriter, r *http.Request, h http.Header, callback func(*http.Request) bool) (*Conn, error) {
	upd := websocket.Upgrader{
		ReadBufferSize:  5120,
		WriteBufferSize: 5120,
		CheckOrigin:     callback,
	}

	conn, err := upd.Upgrade(w, r, h)
	if err != nil {
		return nil, err
	}
	Idx++
	cc := &Conn{
		Idx:     Idx,
		Conn:    conn,
		Addtime: time.Now().Unix(),
	}
	return cc, nil
}

func FindConn(uuid string) ([]*Conn, error) {
	if c, ok := CONNS.Load(uuid); ok {
		conns := c.([]*Conn)
		return conns, nil
	}
	return nil, errors.New(i18n.T("%s 未连接", uuid))
}

func SendNotify(uuid, title, msg, tp string) {
	if c, ok := CONNS.Load(uuid); ok {
		conns := c.([]*Conn)

		bt, _ := json.Marshal(WsResp{Type: "notify", Data: Notify{Title: title, Type: tp, Message: msg}})
		for _, conn := range conns {
			go conn.Conn.WriteMessage(1, bt)
		}
	}
}

func SendMessage(uuid, msg, tp string, plain bool) {
	if c, ok := CONNS.Load(uuid); ok {
		conns := c.([]*Conn)

		bt, _ := json.Marshal(WsResp{Type: "message", Data: Message{Type: tp, Message: msg, Plain: plain}})
		for _, conn := range conns {
			go conn.Conn.WriteMessage(1, bt)
		}
	}
}

func SendContent(uuid, tp string, data any) {
	if c, ok := CONNS.Load(uuid); ok {
		conns := c.([]*Conn)

		bt, _ := json.Marshal(WsResp{Type: tp, Data: data})
		for _, conn := range conns {
			go conn.Conn.WriteMessage(1, bt)
		}
	}
}
