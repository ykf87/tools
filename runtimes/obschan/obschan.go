package obschan

import (
	"context"
	"errors"
	"sync"
)

var ErrCanceled = errors.New("operation canceled")

type ObservableChan struct {
	ch       chan byte
	lock     sync.Mutex
	waitSend int
	waitRecv int
}

func NewObservableChan(buffer int) *ObservableChan {
	return &ObservableChan{
		ch: make(chan byte, buffer),
	}
}

// SendContext 支持 context 取消
func (o *ObservableChan) SendContext(ctx context.Context, b byte) error {
	o.lock.Lock()
	o.waitSend++
	o.lock.Unlock()

	select {
	case o.ch <- b:
		o.lock.Lock()
		o.waitSend--
		o.lock.Unlock()
		return nil
	case <-ctx.Done():
		o.lock.Lock()
		o.waitSend--
		o.lock.Unlock()
		return ErrCanceled
	}
}

// RecvContext 支持 context 取消
func (o *ObservableChan) RecvContext(ctx context.Context) (byte, error) {
	o.lock.Lock()
	o.waitRecv++
	o.lock.Unlock()

	select {
	case b := <-o.ch:
		o.lock.Lock()
		o.waitRecv--
		o.lock.Unlock()
		return b, nil
	case <-ctx.Done():
		o.lock.Lock()
		o.waitRecv--
		o.lock.Unlock()
		return 0, ErrCanceled
	}
}

// Len 返回缓冲区内元素数量
func (o *ObservableChan) Len() int { return len(o.ch) }
func (o *ObservableChan) Cap() int { return cap(o.ch) }

func (o *ObservableChan) WaitingSend() int {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.waitSend
}
func (o *ObservableChan) WaitingRecv() int {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.waitRecv
}
