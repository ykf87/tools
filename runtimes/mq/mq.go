package mq

import (
	"log"
	"sync"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/db/mqs"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var MqClient *MQ

// Handler 定义消息处理函数
type Handler func(msg *mqs.Mq) error

type MQ struct {
	store        *mqs.Store
	handlers     map[string]Handler
	messageChs   map[string]chan *mqs.Mq
	stopCh       chan struct{}
	wg           sync.WaitGroup
	mutex        sync.Mutex
	pollInterval time.Duration
}

func init() {
	MqClient = New(mqs.NewStore(db.MQDB))
}

// New 创建 MQ 实例
func New(store *mqs.Store) *MQ {
	return &MQ{
		store:        store,
		handlers:     make(map[string]Handler),
		messageChs:   make(map[string]chan *mqs.Mq),
		stopCh:       make(chan struct{}),
		pollInterval: 1 * time.Second,
	}
}

// Publish 发布消息
func (mq *MQ) Publish(topic string, payload interface{}, delay time.Duration) (int64, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	id, err := mq.store.Enqueue(topic, string(data), delay)
	if err != nil {
		return 0, err
	}

	// 事件通知：将消息发送到 channel
	mq.mutex.Lock()
	ch, ok := mq.messageChs[topic]
	mq.mutex.Unlock()
	if ok {
		go func() {
			msg, _ := mq.store.FetchByID(id)
			ch <- msg
		}()
	}

	return id, nil
}

// Subscribe 订阅消息
func (mq *MQ) Subscribe(topic string, handler Handler) {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if _, exists := mq.messageChs[topic]; !exists {
		mq.messageChs[topic] = make(chan *mqs.Mq, 100)
	}

	mq.handlers[topic] = handler

	// 启动消费者协程
	mq.wg.Add(1)
	go mq.consumer(topic, mq.messageChs[topic], handler)
}

// consumer 消费逻辑
func (mq *MQ) consumer(topic string, ch chan *mqs.Mq, handler Handler) {
	defer mq.wg.Done()

	// 启动时，先从数据库加载未处理的消息
	for {
		msg, _ := mq.store.FetchPending(topic)
		if msg == nil {
			break
		}
		ch <- msg
	}

	for {
		select {
		case <-mq.stopCh:
			return
		case msg := <-ch:
			if msg == nil {
				time.Sleep(mq.pollInterval)
				continue
			}
			err := handler(msg)
			if err != nil {
				log.Printf("handler error: %v", err)
				mq.store.RetryMessage(msg.ID)
			} else {
				mq.store.MarkDone(msg.ID)
			}
		}
	}
}

// Stop 停止所有消费者
func (mq *MQ) Stop() {
	close(mq.stopCh)
	mq.wg.Wait()
}
