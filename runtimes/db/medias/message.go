package medias

import (
	"time"
	"tools/runtimes/config"
	"tools/runtimes/listens/ws"
)

type MediaDownMessage struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Total   int    `json:"total"`
	Doned   int    `json:"doned"`
	Msg     string `json:"msg"`
	Start   int64  `json:"start"`
	Cover   string `json:"cover"`
	AdminID int64  `json:"-"`
	Status  int    `json:"status"`
}

func (m *Media) Message(total int, cover string) *MediaDownMessage {
	return &MediaDownMessage{
		ID:      m.Id,
		Title:   m.Title,
		Total:   total,
		Doned:   0,
		Msg:     "",
		Start:   time.Now().Unix(),
		Cover:   cover,
		AdminID: m.AdminID,
	}
}

func (msg *MediaDownMessage) Sent(str string, doned int) {
	msg.Msg = str
	msg.Doned = doned
	if bt, err := config.Json.Marshal(map[string]any{
		"type": "",
		"data": msg,
	}); err == nil {
		if msg.AdminID > 0 {
			ws.SentMsg(msg.AdminID, bt)
		} else {
			ws.Broadcost(bt)
		}
	}
}

func (msg *MediaDownMessage) Done() {
	msg.Status = 1
	msg.Sent("下载完成", msg.Total)
}
