package notifies

import (
	"fmt"
	"tools/runtimes/config"
	"tools/runtimes/eventbus"
)

type Notify struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Meta        string `json:"meta"`
	Closeable   bool   `json:"closeable"`
	Url         string `json:"url"`
	Btn         string `json:"btn"`    //点击按钮的文本
	Method      string `json:"method"` // 客户端发起url请求的方法
}

func (t *Notify) Send() {
	if dt, err := config.Json.Marshal(map[string]any{"type": "notify", "data": t}); err == nil {
		eventbus.Bus.Publish("notify", dt)
	} else {
		fmt.Println("message格式化错误!")
	}
}
