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

type TaskObj interface {
	Closer()
}

type TaskRow struct {
	Module string  `json:"module"` //模型,根据这个进行分组
	Key    string  `json:"key"`    //识别码,跳转任务的关键词
	Title  string  `json:"title"`  //标题,显示的任务名称
	To     TaskObj `json:"-"`      //任务的接口
}

var CONNS sync.Map
var TasksMap sync.Map

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

func FindConn(uuid string) ([]*Conn, error) {
	if c, ok := CONNS.Load(uuid); ok {
		conns := c.([]*Conn)
		return conns, nil
	}
	return nil, errors.New(i18n.T("%s 未连接", uuid))
}

func SendNotify(uuid, title, msg, tp string) {
	if uuid == "" {
		CONNS.Range(func(k, v any) bool {
			conns := v.([]*Conn)
			bt, _ := json.Marshal(WsResp{Type: "notify", Data: Notify{Title: title, Type: tp, Message: msg}})
			for _, conn := range conns {
				go conn.WriteMessage(bt)
			}
			return true
		})
	} else if c, ok := CONNS.Load(uuid); ok {
		conns := c.([]*Conn)

		bt, _ := json.Marshal(WsResp{Type: "notify", Data: Notify{Title: title, Type: tp, Message: msg}})
		for _, conn := range conns {
			go conn.WriteMessage(bt)
			// go conn.Conn.WriteMessage(1, bt)
		}
	}
}

func SendMessage(uuid, msg, tp string, plain bool) {
	if uuid == "" {
		CONNS.Range(func(k, v any) bool {
			conns := v.([]*Conn)
			bt, _ := json.Marshal(WsResp{Type: "message", Data: Message{Type: tp, Message: msg, Plain: plain}})
			for _, conn := range conns {
				go conn.WriteMessage(bt)
			}
			return true
		})
	} else if c, ok := CONNS.Load(uuid); ok {
		conns := c.([]*Conn)

		bt, _ := json.Marshal(WsResp{Type: "message", Data: Message{Type: tp, Message: msg, Plain: plain}})
		for _, conn := range conns {
			go conn.WriteMessage(bt)
			// go conn.Conn.WriteMessage(1, bt)
		}
	}
}

func SendContent(uuid, tp string, data any) {
	if uuid == "" {
		CONNS.Range(func(k, v any) bool {
			conns := v.([]*Conn)
			bt, _ := json.Marshal(WsResp{Type: tp, Data: data})
			for _, conn := range conns {
				go conn.WriteMessage(bt)
			}
			return true
		})
	} else if c, ok := CONNS.Load(uuid); ok {
		conns := c.([]*Conn)

		bt, _ := json.Marshal(WsResp{Type: tp, Data: data})
		for _, conn := range conns {
			go conn.WriteMessage(bt)
		}
	}
}

type TaskResp struct {
	Name  string     `json:"name"`
	Tasks []*TaskRow `json:"tasks"`
}

func SendTask() {
	var dtts []*TaskResp
	TasksMap.Range(func(k, v any) bool {
		tr := new(TaskResp)
		tr.Name = k.(string)
		tr.Tasks = v.([]*TaskRow)

		dtts = append(dtts, tr)
		return true
	})
	dt, _ := json.Marshal(WsResp{Type: "task", Data: dtts})
	CONNS.Range(func(k, v any) bool {
		// uuid := k.(string)
		conns := v.([]*Conn)

		for _, conn := range conns {
			go conn.WriteMessage(dt)
		}
		return true
	})
}

//发送版本更新
func Upgrade(data any) {
	CONNS.Range(func(k, v any) bool {
		conns := v.([]*Conn)
		bt, _ := json.Marshal(WsResp{Type: "upgrade", Data: data})
		for _, conn := range conns {
			go conn.WriteMessage(bt)
			break
		}
		return true
	})
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

//链接默认发送的信息
func SendDefault() {
	SendTask()
}
