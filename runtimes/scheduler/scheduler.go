package scheduler

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Scheduler struct {
	mu sync.Mutex

	Ctx    context.Context
	cancel context.CancelFunc

	tasks   map[string]*Task // waiting / failed
	running map[string]*Task // running
	heap    taskHeap         // waiting heap
	sem     chan struct{}    // global concurrency
	wakeup  chan struct{}
	StopCh  chan struct{}

	logs    []ExecLog
	maxLogs int

	opts Options
}

func New(opts Options) *Scheduler {
	if opts.MaxConcurrency <= 0 {
		opts.MaxConcurrency = 1
	}
	if opts.MaxQueueSize <= 0 {
		opts.MaxQueueSize = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Scheduler{
		Ctx:     ctx,
		cancel:  cancel,
		tasks:   make(map[string]*Task),
		running: make(map[string]*Task),
		heap:    taskHeap{},
		sem:     make(chan struct{}, opts.MaxConcurrency),
		wakeup:  make(chan struct{}, 1),
		StopCh:  make(chan struct{}),
		opts:    opts,
		logs:    make([]ExecLog, 0, 1000),
		maxLogs: 1000,
	}

	heap.Init(&s.heap)
	s.load()

	return s
}

func (s *Scheduler) Start() {
	go s.loop()
}

func (s *Scheduler) Stop() {
	s.cancel()

	s.mu.Lock()
	for _, t := range s.running {
		t.Canceled = true
		if t.Stop != nil {
			go t.Stop(s.Ctx)
		}
	}
	s.mu.Unlock()

	close(s.StopCh)
	s.save()
}

func (s *Scheduler) Exists(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; ok {
		return true
	}
	if _, ok := s.running[id]; ok {
		return true
	}
	return false
}

func (s *Scheduler) Add(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t.Run == nil {
		return fmt.Errorf("Run 必须传入")
	}
	if len(s.tasks) >= s.opts.MaxQueueSize {
		return errors.New("task queue full")
	}
	if _, ok := s.tasks[t.ID]; ok {
		return errors.New("task already exists")
	}

	if t.mutex == nil {
		t.mutex = make(chan struct{}, 1)
	}

	t.Status = TaskWaiting
	t.retryCount = 0
	t.Canceled = false
	t.NextRun = time.Now()
	t.index = -1

	s.tasks[t.ID] = t
	heap.Push(&s.heap, t)
	s.signal()

	go s.save()
	return nil
}

func (s *Scheduler) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	closed := false
	if t, ok := s.tasks[id]; ok {
		t.Canceled = true
		if t.index >= 0 {
			heap.Remove(&s.heap, t.index)
		}
		if t.Stop != nil {
			closed = true
			go t.Stop(nil)
		}
		delete(s.tasks, id)
	}

	if t, ok := s.running[id]; ok {
		t.Canceled = true
		if !closed {
			go t.Stop(nil)
		}
	}

	s.save()
}

func (s *Scheduler) Pause(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.tasks[id]; ok {
		t.Paused = true
	}
	s.save()
}

func (s *Scheduler) Resume(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.tasks[id]; ok {
		t.Paused = false
		s.signal()
	}
	s.save()
}

func (s *Scheduler) loop() {
	for {
		s.mu.Lock()
		for len(s.heap) == 0 {
			s.mu.Unlock()
			select {
			case <-s.wakeup:
				s.mu.Lock()
			case <-s.StopCh:
				return
			}
		}

		t := s.heap[0]

		if t.Canceled {
			heap.Pop(&s.heap)
			delete(s.tasks, t.ID)
			s.mu.Unlock()
			continue
		}

		wait := time.Until(t.NextRun)
		s.mu.Unlock()

		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-s.wakeup:
				timer.Stop()
				continue
			case <-s.StopCh:
				timer.Stop()
				return
			}
		}

		s.mu.Lock()
		heap.Pop(&s.heap)
		delete(s.tasks, t.ID)
		s.mu.Unlock()

		if !t.Paused && !t.Canceled {
			go s.runTask(t)
		}
	}
}

func (s *Scheduler) runTask(t *Task) {
	// ===== 全局并发控制 =====
	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	// ===== 单任务互斥（这是执行屏障）=====
	t.mutex <- struct{}{}
	defer func() { <-t.mutex }()

	// ⭐⭐⭐ 关键：在真正执行前，最后一次确认是否被删除
	if t.Canceled {
		return
	}

	start := time.Now()
	s.markRunning(t)

	defer func() {
		if r := recover(); r != nil {
			s.handleFail(t, fmt.Errorf("panic: %v", r), start)
			s.finishTask(t)
		}
	}()

	ctx := context.Background()
	var cancel context.CancelFunc
	if t.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, t.Timeout)
		defer cancel()
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- t.Run(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			s.handleFail(t, err, start)
		} else {
			s.handleSuccess(t, start)
		}
	case <-ctx.Done():
		s.handleFail(t, ctx.Err(), start)
	}

	s.finishTask(t)
}

func (s *Scheduler) markRunning(t *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t.Status = TaskRunning
	t.LastRun = time.Now()
	s.running[t.ID] = t
}

func (s *Scheduler) handleSuccess(t *Task, start time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t.Status = TaskSuccess
	t.retryCount = 0
	t.LastError = ""
	t.FinishTime = time.Now()

	s.appendLogLocked(ExecLog{
		TaskID:     t.ID,
		StartTime:  start,
		EndTime:    t.FinishTime,
		Duration:   time.Since(start),
		Success:    true,
		RetryCount: t.retryCount,
	})
}

func (s *Scheduler) handleFail(t *Task, err error, start time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t.LastError = err.Error()
	t.FinishTime = time.Now()

	timeout := errors.Is(err, context.DeadlineExceeded)

	if t.retryCount < t.MaxRetry && !t.Canceled {
		t.retryCount++
		t.Status = TaskWaiting
		t.NextRun = time.Now().Add(t.RetryGap)
	} else {
		t.Status = TaskFailed
	}

	s.appendLogLocked(ExecLog{
		TaskID:     t.ID,
		StartTime:  start,
		EndTime:    t.FinishTime,
		Duration:   time.Since(start),
		Success:    false,
		Error:      err.Error(),
		Timeout:    timeout,
		RetryCount: t.retryCount,
	})
}

func (s *Scheduler) finishTask(t *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.running, t.ID)

	if t.Canceled {
		return
	}

	if t.Status == TaskWaiting {
		t.index = -1
		s.tasks[t.ID] = t
		heap.Push(&s.heap, t)
		s.signal()
		return
	}

	if t.Status == TaskSuccess && t.Interval > 0 {
		t.Status = TaskWaiting
		t.NextRun = time.Now().Add(t.Interval)
		t.index = -1
		s.tasks[t.ID] = t
		heap.Push(&s.heap, t)
		s.signal()
		return
	}

	if t.Status == TaskFailed {
		s.tasks[t.ID] = t
	}
}

func (s *Scheduler) appendLogLocked(l ExecLog) {
	if len(s.logs) >= s.maxLogs {
		s.logs = s.logs[1:]
	}
	s.logs = append(s.logs, l)
}

func (s *Scheduler) signal() {
	select {
	case s.wakeup <- struct{}{}:
	default:
	}
}

func (s *Scheduler) RunningTasks() []*Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := make([]*Task, 0, len(s.running))
	for _, t := range s.running {
		list = append(list, t)
	}
	return list
}

func (s *Scheduler) WaitingTasks() []*Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := make([]*Task, 0, len(s.heap))
	for _, t := range s.heap {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].NextRun.Before(list[j].NextRun)
	})
	return list
}

func (s *Scheduler) FailedTasks() []*Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res []*Task
	for _, t := range s.tasks {
		if t.Status == TaskFailed {
			res = append(res, t)
		}
	}
	return res
}

func (s *Scheduler) Logs() []ExecLog {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]ExecLog, len(s.logs))
	copy(cp, s.logs)
	return cp
}
