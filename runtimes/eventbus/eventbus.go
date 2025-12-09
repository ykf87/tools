package eventbus

import (
	"sync"
)

// Handler 定义事件处理函数
type Handler func(data interface{})

// EventBus 是事件总线
type EventBus struct {
	mutex    sync.RWMutex
	handlers map[string][]Handler // topic -> handler 列表
}

var Bus *EventBus

func init() {
	Bus = New()
}

// New 创建 EventBus 实例
func New() *EventBus {
	return &EventBus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(topic string, handler Handler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	eb.handlers[topic] = append(eb.handlers[topic], handler)
}

// Unsubscribe 移除某个 handler（可选）
func (eb *EventBus) Unsubscribe(topic string, handler Handler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	list := eb.handlers[topic]
	for i, h := range list {
		if &h == &handler { // 这里比较指针即可
			eb.handlers[topic] = append(list[:i], list[i+1:]...)
			break
		}
	}
}

// Publish 发布事件
func (eb *EventBus) Publish(topic string, data any) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	if list, ok := eb.handlers[topic]; ok {
		for _, handler := range list {
			// 异步调用 handler，避免阻塞
			go handler(data)
		}
	}
}
