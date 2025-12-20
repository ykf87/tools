package clients

import (
	"sync"
	"tools/runtimes/eventbus"
	"tools/runtimes/ws"

	"github.com/tidwall/gjson"
)

type WsConn struct {
	Conn     *ws.Conn
	PhoneId  int64
	DeviceId string
	Send     chan []byte // 发送队列
	Groups   map[string]bool
}

type Hub struct {
	mu sync.RWMutex

	Clients map[string]*WsConn
	Groups  map[string]map[string]*WsConn // group => clientID => client
}

var Hubs *Hub

func init() {
	Hubs = &Hub{}
	Hubs.Clients = make(map[string]*WsConn)
	Hubs.Groups = make(map[string]map[string]*WsConn)
}

// 监听消息
func (this *WsConn) listens() {
	for {
		p, err := this.Conn.ReadMessage()
		if err != nil {
			break
		}
		ps := gjson.ParseBytes(p)
		eventbus.Bus.Publish(ps.Get("tp").String(), ps.Get("data").String())
	}
}

// 注册
func (this *Hub) Register(dbid int64, devid string, conn *ws.Conn) {
	this.mu.Lock()
	defer this.mu.Unlock()
	nc := &WsConn{
		PhoneId:  dbid,
		DeviceId: devid,
		Conn:     conn,
	}
	nc.Send = make(chan []byte)
	nc.Groups = make(map[string]bool)
	go nc.listens()
	this.Clients[devid] = nc
}

// 单发给某个客户端
func (h *Hub) SentClient(uuid string, data []byte) {
	if c, ok := h.Clients[uuid]; ok {
		if c.Conn != nil {
			c.Conn.WriteMessage(data)
		}
	}
}

// 添加到组
func (h *Hub) JoinGroup(groupID, deviceid string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.Groups[groupID]; !ok {
		h.Groups[groupID] = make(map[string]*WsConn)
	}

	client, ok := h.Clients[deviceid]
	if ok {
		h.Groups[groupID][deviceid] = client
		client.Groups[groupID] = true
	}
}

// 群组发送消息
func (h *Hub) SendToGroup(groupID string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if group, ok := h.Groups[groupID]; ok {
		for _, client := range group {
			if client.Conn != nil {
				client.Conn.WriteMessage(msg)
			}
		}
	}
}

// 广播消息, fun 为广播条件
func (h *Hub) Broadcast(msg []byte, fun func(*WsConn) bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		if fun != nil && fun(client) == false {
			continue
		}
		if client.Conn != nil {
			client.Conn.WriteMessage(msg)
		}
	}
}

// 关闭某个连接
func (h *Hub) Close(deviceId string) {
	if cli, ok := h.Clients[deviceId]; ok {
		for gn, ok := range cli.Groups {
			if ok {
				if gps, ok := h.Groups[gn]; ok {
					if _, ok := gps[deviceId]; ok {
						delete(h.Groups[gn], deviceId)
					}
				}
			}
		}
		cli.Conn.Close()
		delete(h.Clients, deviceId)
	}
}
