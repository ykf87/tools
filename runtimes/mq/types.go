package mq

import "context"

type Message struct {
	ID      int64
	Topic   string
	Payload string

	Retry int
}

type Handler func(ctx context.Context, msg *Message) error
