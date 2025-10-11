package ws

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
)

// var CONNS sync.Map
// var TasksMap sync.Map

var Idx int64

type Conn struct {
	Conn      *websocket.Conn
	Idx       int64
	Addtime   int64
	InChan    chan []byte
	OutChan   chan []byte
	CloseChan chan byte
	Mutex     sync.Mutex
	IsClose   bool
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
		Idx:       Idx,
		Conn:      conn,
		Addtime:   time.Now().Unix(),
		InChan:    make(chan []byte, 1024),
		OutChan:   make(chan []byte, 1024),
		CloseChan: make(chan byte, 1),
	}

	go cc.reader()
	go cc.writer()
	return cc, nil
}

// 消息读取
func (conn *Conn) reader() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.Conn.ReadMessage(); err != nil {
			goto ERR
		}
		select {
		case <-conn.CloseChan:
			goto ERR
		case conn.InChan <- data:
		}

	}
ERR:
	conn.Close()
}

// 发送消息
func (conn *Conn) writer() {
	var (
		data []byte
		err  error
	)
	for {
		select {
		case <-conn.CloseChan:
			goto ERR
		case data = <-conn.OutChan:
		}

		if conn.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}

// 线程安全的关闭
func (conn *Conn) Close() {
	conn.Conn.Close()

	conn.Mutex.Lock()
	if !conn.IsClose {
		close(conn.CloseChan)
		conn.IsClose = true
	}
	conn.Mutex.Unlock()
}

// 线程安全的消息读取
func (conn *Conn) ReadMessage() (data []byte, err error) {
	select {
	case <-conn.CloseChan:
		err = errors.New("链接被关闭")
	case data = <-conn.InChan:
	}
	return
}

// 线程安全的消息写入
func (conn *Conn) WriteMessage(data []byte) (err error) {
	select {
	case conn.OutChan <- data:
	case <-conn.CloseChan:
		err = errors.New("链接被关闭!")
	}
	return
}
