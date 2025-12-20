package mq

import (
	"context"
	"sync"
)

type MQ struct {
	store Store

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.Mutex

	started bool

	queues   map[string]chan *Message
	handlers map[string]Handler

	maxRetry int
}

func New(store Store, maxRetry int) *MQ {
	ctx, cancel := context.WithCancel(context.Background())

	return &MQ{
		store:    store,
		ctx:      ctx,
		cancel:   cancel,
		queues:   make(map[string]chan *Message),
		handlers: make(map[string]Handler),
		maxRetry: maxRetry,
	}
}

func (mq *MQ) Register(topic string, handler Handler, buffer int) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if _, ok := mq.handlers[topic]; ok {
		panic("duplicate topic: " + topic)
	}

	ch := make(chan *Message, buffer)
	mq.handlers[topic] = handler
	mq.queues[topic] = ch

	if mq.started {
		mq.wg.Add(1)
		go mq.consumeLoop(topic, ch, handler)
	}
}

func (mq *MQ) Start() error {
	mq.mu.Lock()
	if mq.started {
		mq.mu.Unlock()
		return nil
	}
	mq.started = true
	mq.mu.Unlock()

	msgs, err := mq.store.LoadPending()
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		if ch, ok := mq.queues[msg.Topic]; ok {
			ch <- msg
		}
	}

	for topic, handler := range mq.handlers {
		ch := mq.queues[topic]
		mq.wg.Add(1)
		go mq.consumeLoop(topic, ch, handler)
	}

	return nil
}

func (mq *MQ) Publish(topic, payload string) error {
	id, err := mq.store.Save(topic, payload)
	if err != nil {
		return err
	}

	msg := &Message{
		ID:      id,
		Topic:   topic,
		Payload: payload,
	}

	mq.mu.Lock()
	ch, ok := mq.queues[topic]
	mq.mu.Unlock()

	if ok {
		ch <- msg
	}
	return nil
}

func (mq *MQ) Stop() {
	mq.cancel()
	mq.wg.Wait()
}
