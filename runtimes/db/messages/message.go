package messages

import (
	"fmt"
	"tools/runtimes/config"
	"tools/runtimes/eventbus"
)

type Message struct {
	Type     string `json:"type"`
	Content  string `json:"content"`
	Duration int64  `json:"duration"`
}

func (t *Message) Send() {
	if dt, err := config.Json.Marshal(map[string]any{"type": "message", "data": t}); err == nil {
		eventbus.Bus.Publish("message", dt)
	} else {
		fmt.Println("message格式化错误!")
	}
}
