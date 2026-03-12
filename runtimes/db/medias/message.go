package medias

import (
	"time"
	"tools/runtimes/config"
	"tools/runtimes/listens/ws"
	"tools/runtimes/storage"
)

type MediaResponseMessage struct {
	Type   int   `json:"type"` // 0媒体文件 1目录 2下载中的
	PathID int64 `json:"path_id"`

	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Platform string `json:"platform"`

	Total   int    `json:"total"`
	Doned   int    `json:"doned"`
	Msg     string `json:"msg"`
	Start   int64  `json:"start"`
	Status  int    `json:"status"`
	AdminID int64  `json:"-"`

	Cover        string       `json:"cover"`
	CoverStorage string       `json:"-"`
	Sizes        int64        `json:"sizes"`
	Numbers      int64        `json:"numbers"`
	Files        []*MediaFile `json:"files"`
	User         *MediaUser   `json:"user"`

	Addtime time.Time `json:"addtime"`
}

func (mms MediaResponseMessage) MarshalJSON() ([]byte, error) {
	type Alias MediaResponseMessage

	a := Alias(mms)
	if a.Cover != "" {
		a.Cover = storage.Load(a.CoverStorage).URL(a.Cover)
	}
	// if a.User != nil {
	// 	if a.User.Cover != "" {
	// 		if a.User.CoverStorage == "" {
	// 			a.User.CoverStorage = "local"
	// 		}
	// 		a.User.Cover = storage.Load(a.User.CoverStorage).URL(a.User.Cover)
	// 	}
	// }
	// for _, v := range a.Files {
	// 	if v.FileName != "" {
	// 		v.FileName = storage.Load(v.FileSystem).URL(v.FileName)
	// 	}
	// }

	return config.Json.Marshal(a)
}

func (m *Media) Message(total int, cover string, pathID int64) *MediaResponseMessage {
	if m.User != nil {
		m.User.Cover = storage.Load("").URL(m.User.Cover)
	}
	return &MediaResponseMessage{
		ID:           m.Id,
		Title:        m.Title,
		Total:        total,
		Doned:        0,
		Msg:          "",
		Start:        time.Now().Unix(),
		Cover:        cover,
		CoverStorage: m.CoverStorage,
		AdminID:      m.AdminID,
		Addtime:      time.Now(),
		User:         m.User,
		Type:         2,
		Platform:     m.Platform,
		PathID:       pathID,
	}
}

func (msg *MediaResponseMessage) Sent(str string, doned int) {
	msg.Msg = str
	msg.Doned = doned
	if bt, err := config.Json.Marshal(map[string]any{
		"type": "media-download",
		"data": msg,
	}); err == nil {
		if msg.AdminID > 0 {
			ws.SentMsg(msg.AdminID, bt)
		} else {
			ws.Broadcost(bt)
		}
	}
}

func (msg *MediaResponseMessage) Done(str string) {
	if str == "" {
		str = "下载完成"
	}
	msg.Status = 1
	msg.Sent(str, msg.Total)
}
