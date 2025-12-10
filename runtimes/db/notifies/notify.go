package notifies

import (
	"fmt"
	"tools/runtimes/config"
	"tools/runtimes/eventbus"
)

type Notify struct { // 不需要存数据库,因为通知完就结束了
	Type        string `json:"type" gorm:"index;not null"`
	Title       string `json:"title" gorn:"default:null;type:varchar(60)"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Meta        string `json:"meta" gorm:"-"`                              //
	Closeable   bool   `json:"closeable" gorm:"type:tinyint(1);default:0"` // 是否可关闭
	Url         string `json:"url" gorm:"default:null"`                    // 点击的请求连接
	Btn         string `json:"btn"`                                        // 点击按钮的文本
	Method      string `json:"method"`                                     // 客户端发起url请求的方法
}

func (t *Notify) Send() {
	if dt, err := config.Json.Marshal(map[string]any{"type": "notify", "data": t}); err == nil {
		eventbus.Bus.Publish("notify", dt)
	} else {
		fmt.Println("message格式化错误!")
	}
}
