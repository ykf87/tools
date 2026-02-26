package expire

import (
	"container/heap"
	"context"
	"sync"
	"time"
	"tools/runtimes/mainsignal"
)

type task struct {
	expireTs int64
	callback func()
	index    int
}

type taskHeap []*task

var ssc *Scheduler

func init() {
	ssc = new(mainsignal.MainCtx)
}

func (h taskHeap) Len() int { return len(h) }

func (h taskHeap) Less(i, j int) bool {
	return h[i].expireTs < h[j].expireTs
}

func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *taskHeap) Push(x any) {
	t := x.(*task)
	t.index = len(*h)
	*h = append(*h, t)
}

func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	t := old[n-1]
	old[n-1] = nil
	t.index = -1
	*h = old[:n-1]
	return t
}

type Scheduler struct {
	mu     sync.Mutex
	tasks  taskHeap
	wakeup chan struct{}
	ctx    context.Context
}

func new(ctx context.Context) *Scheduler {
	s := &Scheduler{
		tasks:  make(taskHeap, 0),
		wakeup: make(chan struct{}, 1),
		ctx:    ctx,
	}

	heap.Init(&s.tasks)
	go s.loop()

	return s
}

func Add(expireTs int64, cb func()) {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()

	t := &task{
		expireTs: expireTs,
		callback: cb,
	}

	heap.Push(&ssc.tasks, t)
	ssc.notify()
}

func (s *Scheduler) loop() {
	var timer *time.Timer

	for {
		s.mu.Lock()

		if len(s.tasks) == 0 {
			s.mu.Unlock()

			select {
			case <-s.wakeup:
				continue
			case <-s.ctx.Done():
				return
			}
		}

		next := s.tasks[0]
		nowTs := time.Now().Unix()

		// 已到期
		if next.expireTs <= nowTs {
			heap.Pop(&s.tasks)
			s.mu.Unlock()

			go next.callback()
			continue
		}

		wait := time.Duration(next.expireTs-nowTs) * time.Second

		if timer == nil {
			timer = time.NewTimer(wait)
		} else {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(wait)
		}

		s.mu.Unlock()

		select {
		case <-timer.C:
		case <-s.wakeup:
		case <-s.ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return
		}
	}
}

func (s *Scheduler) notify() {
	select {
	case s.wakeup <- struct{}{}:
	default:
	}
}
