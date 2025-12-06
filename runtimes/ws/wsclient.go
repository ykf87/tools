package ws

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type EventHandler struct {
	OnOpen    func()
	OnMessage func(messageType int, message []byte)
	OnError   func(err error)
	OnClose   func()
	ConnHead  func() http.Header
}

type Config struct {
	Proxy         string
	RetryInterval time.Duration
	MaxRetry      int           // -1 表示无限次
	PingInterval  time.Duration // 使用 WebSocket 原生 Ping
	PingMessage   []byte        // ping发送的内容
	WriteWait     time.Duration // 写超时
	ReadWait      time.Duration // Pong 超时
	QueueSize     int           // 重连期间的消息缓存大小
}

type Client struct {
	addr    string
	conn    *websocket.Conn
	dialer  *websocket.Dialer
	handler EventHandler

	cfg     Config
	mu      sync.Mutex
	closing bool

	sendQueue chan []byte
}

func New(addr string, cfg *Config, handler EventHandler) *Client {
	// pc, file, line, ok := runtime.Caller(1)
	// if ok {
	// 	fmt.Printf("Called from %s:%d (%s)\n", file, line, runtime.FuncForPC(pc).Name())
	// }
	def := Config{
		RetryInterval: 3 * time.Second,
		MaxRetry:      -1,
		PingInterval:  25 * time.Second,
		ReadWait:      60 * time.Second,
		WriteWait:     10 * time.Second,
		QueueSize:     256,
		PingMessage:   nil,
	}

	if cfg != nil {
		if cfg.RetryInterval > 0 {
			def.RetryInterval = cfg.RetryInterval
		}
		if cfg.MaxRetry != 0 {
			def.MaxRetry = cfg.MaxRetry
		}
		if cfg.PingInterval > 0 {
			def.PingInterval = cfg.PingInterval
		}
		if cfg.ReadWait > 0 {
			def.ReadWait = cfg.ReadWait
		}
		if cfg.WriteWait > 0 {
			def.WriteWait = cfg.WriteWait
		}
		if cfg.QueueSize > 0 {
			def.QueueSize = cfg.QueueSize
		}
		if cfg.Proxy != "" {
			// u, err := url.Parse(cfg.Proxy)
			if _, err := url.Parse(cfg.Proxy); err == nil {
				def.Proxy = cfg.Proxy
				// log.Println("[wsclient] 使用代理:", cfg.Proxy)
			}
		}
		if cfg.PingMessage != nil {
			def.PingMessage = cfg.PingMessage
		}
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	if def.Proxy != "" {
		if proxyURL, err := url.Parse(def.Proxy); err == nil {
			dialer.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &Client{
		addr:      addr,
		dialer:    dialer,
		cfg:       def,
		handler:   handler,
		sendQueue: make(chan []byte, def.QueueSize),
	}
}

func (c *Client) Start() {
	go c.connectLoop()
}

func (c *Client) connectLoop() {
	retries := 0

	for {
		if c.closing {
			return
		}

		err := c.connect()
		if err != nil {
			if c.handler.OnError != nil {
				c.handler.OnError(err)
			}
			// log.Println("[wsclient] 连接失败:", err)

			retries++
			if c.cfg.MaxRetry != -1 && retries > c.cfg.MaxRetry {
				// log.Println("[wsclient] 达到最大重连次数，停止重连")
				return
			}

			time.Sleep(c.cfg.RetryInterval)
			continue
		}

		retries = 0
		c.run() // 阻塞直到连接断开

		if c.handler.OnClose != nil {
			c.handler.OnClose()
		}

		if c.closing {
			return
		}

		time.Sleep(c.cfg.RetryInterval)
	}
}

func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var hd http.Header
	if c.handler.ConnHead != nil {
		hd = c.handler.ConnHead()
	}
	conn, _, err := c.dialer.Dial(c.addr, hd)
	if err != nil {
		return err
	}

	conn.SetReadDeadline(time.Now().Add(c.cfg.ReadWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(c.cfg.ReadWait))
		return nil
	})

	c.conn = conn

	if c.handler.OnOpen != nil {
		c.handler.OnOpen()
	}

	return nil
}

func (c *Client) run() {
	done := make(chan struct{})

	go c.readLoop(done)
	go c.writeLoop(done)
	go c.pingLoop(done)

	<-done // 任一退出，代表连接断开

	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()
}

func (c *Client) Send(msg []byte) error {
	if c.closing {
		return errors.New("client closing")
	}

	select {
	case c.sendQueue <- msg:
		return nil
	default:
		return errors.New("send queue full")
	}
}

func (c *Client) writeLoop(done chan struct{}) {
	for {
		select {
		case msg := <-c.sendQueue:
			c.mu.Lock()
			if c.conn == nil {
				c.mu.Unlock()
				continue
			}

			c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteWait))
			err := c.conn.WriteMessage(websocket.TextMessage, msg)
			c.mu.Unlock()

			if err != nil {
				if c.handler.OnError != nil {
					c.handler.OnError(err)
				}
				close(done)
				return
			}
		case <-done:
			return
		}
	}
}

func (c *Client) readLoop(done chan struct{}) {
	for {
		// c.mu.Lock()
		if c.conn == nil {
			// c.mu.Unlock()
			break
		}
		mt, msg, err := c.conn.ReadMessage()
		if mt == websocket.PongMessage {
			fmt.Println("客户端接收到pong------")
			continue
		}
		// c.mu.Unlock()

		if err != nil {
			if c.handler.OnError != nil {
				c.handler.OnError(err)
			}
			close(done)
			return
		}

		if c.handler.OnMessage != nil {
			c.handler.OnMessage(mt, msg)
		}
	}
}

func (c *Client) pingLoop(done chan struct{}) {
	ticker := time.NewTicker(c.cfg.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// c.Send(c.cfg.PingMessage)
			c.mu.Lock()
			if c.conn != nil {
				c.conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, c.cfg.PingMessage); err != nil {
					c.mu.Unlock()
					close(done)
					return
				}
			}
			c.mu.Unlock()
		case <-done:
			return
		}
	}
}

func (c *Client) Close() {
	c.closing = true

	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()

	close(c.sendQueue)
}
