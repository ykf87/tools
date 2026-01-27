package scheduler

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu   sync.Mutex
	pq   runnerHeap
	wake chan struct{}
}

func New() *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		ctx:    ctx,
		cancel: cancel,
		wake:   make(chan struct{}, 1),
	}
	heap.Init(&s.pq)
	go s.loop()
	return s
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.signal()
}

func (s *Scheduler) NewRunner(task TaskFunc, timeout time.Duration) *Runner {
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(s.ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(s.ctx)
	}

	r := newRunner(ctx, cancel, task, s)
	r.id = uuid.NewString()
	return r
}

func (s *Scheduler) enqueue(r *Runner) {
	s.mu.Lock()
	heap.Push(&s.pq, r)
	s.mu.Unlock()
	s.signal()
}

func (s *Scheduler) signal() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Scheduler) loop() {
	for {
		s.mu.Lock()
		next := s.pq.Peek()
		s.mu.Unlock()

		if next == nil {
			select {
			case <-s.wake:
				continue
			case <-s.ctx.Done():
				return
			}
		}

		wait := time.Until(next.nextRun)
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-s.wake:
				timer.Stop()
				continue
			case <-s.ctx.Done():
				timer.Stop()
				return
			}
		}

		s.mu.Lock()
		r := heap.Pop(&s.pq).(*Runner)
		s.mu.Unlock()

		go r.execute()
	}
}
