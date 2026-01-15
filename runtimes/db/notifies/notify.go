package notifies

import (
	"tools/runtimes/eventbus"
)

var notifyID int64

type Notify struct { // 不需要存数据库,因为通知完就结束了
	ID          int64   `json:"id"`
	Type        string  `json:"type" gorm:"index;not null"`
	AdminID     int64   `json:"admin_id"`
	Title       string  `json:"title" gorn:"default:null;type:varchar(60)"`
	Description string  `json:"description"`
	Content     string  `json:"content"`
	Schedule    float64 `json:"schedule"`                                   // 进度
	Meta        string  `json:"meta" gorm:"-"`                              // 参考native ui的notify的meta
	KeepOpen    bool    `json:"keep_open" gorm:"type:tinyint(1);default:0"` // 是否可关闭
	Url         string  `json:"url" gorm:"default:null"`                    // 点击的请求连接
	Btn         string  `json:"btn"`                                        // 点击按钮的文本
	Method      string  `json:"method"`                                     // 客户端发起url请求的方法
	Done        bool    `json:"done"`                                       // true表示结束了
}

func NewNotify() *Notify {
	notifyID++
	return &Notify{
		ID: notifyID,
	}
}

func (t *Notify) Send() {
	if t.ID < 1 {
		notifyID++
		t.ID = notifyID
	}
	eventbus.Bus.Publish("notify", t)
	// if dt, err := config.Json.Marshal(map[string]any{"type": "notify", "data": t}); err == nil {
	// 	eventbus.Bus.Publish("notify", dt)
	// } else {
	// 	fmt.Println("message格式化错误!")
	// }
}
