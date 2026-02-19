package sch

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

type TaskType int

const (
	TaskInterval TaskType = iota
	TaskCron
)

type Scheduler struct {
	cron           *cron.Cron
	maxConcurrency int32
	semaphore      chan struct{}
	tasks          sync.Map
	stopped        int32
}

type Runner struct {
	id       string
	taskType TaskType

	interval time.Duration
	cronExpr string
	entryID  cron.EntryID

	timeout    time.Duration
	retry      int
	retryDelay time.Duration

	jitter float64 // 0.2 = ±20%

	job func(context.Context) error

	onComplete func(id string, err error)
	onClose    func(id string)

	runCount        int64
	totalRetryCount int64
	lastRetryCount  int64
	nextRun         atomic.Value // time.Time
	expireAt        time.Time

	ctx    context.Context
	cancel context.CancelFunc

	running int32
	closed  int32

	s *Scheduler
}

func NewScheduler(maxConcurrency int) *Scheduler {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	return &Scheduler{
		cron:           cron.New(cron.WithSeconds()),
		maxConcurrency: int32(maxConcurrency),
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	if !atomic.CompareAndSwapInt32(&s.stopped, 0, 1) {
		return
	}

	s.tasks.Range(func(key, value any) bool {
		r := value.(*Runner)
		r.close()
		return true
	})

	ctx := s.cron.Stop()
	<-ctx.Done()
}

func (s *Scheduler) tryAcquire() bool {
	select {
	case s.semaphore <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Scheduler) release() {
	select {
	case <-s.semaphore:
	default:
	}
}

func (r *Runner) GetID() string {
	return r.id
}

func (r *Runner) GetCtx() context.Context {
	return r.ctx
}

func (r *Runner) RunCount() int64 {
	return atomic.LoadInt64(&r.runCount)
}

func (r *Runner) NextRunTime() time.Time {
	v := r.nextRun.Load()
	if v == nil {
		return time.Time{}
	}
	return v.(time.Time)
}

func (r *Runner) SetJitter(percent float64) {
	if percent >= 0 && percent <= 1 {
		r.jitter = percent
	}
}

func applyJitter(base time.Time, percent float64) time.Time {
	if percent <= 0 {
		return base
	}
	delta := time.Until(base)
	if delta <= 0 {
		return base
	}

	offsetRange := float64(delta) * percent
	offset := (rand.Float64()*2 - 1) * offsetRange
	return base.Add(time.Duration(offset))
}

func (r *Runner) calcNextIntervalRun() time.Time {
	base := time.Now().Add(r.interval)
	return applyJitter(base, r.jitter)
}

func (r *Runner) calcNextCronRun(base time.Time) time.Time {
	return applyJitter(base, r.jitter)
}

func (r *Runner) execute() {

	if atomic.LoadInt32(&r.closed) == 1 {
		return
	}

	if !atomic.CompareAndSwapInt32(&r.running, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&r.running, 0)

	if !r.s.tryAcquire() {
		return
	}
	defer r.s.release()

	var finalErr error

	var retryTimes int64 = 0
	for i := 0; i <= r.retry; i++ {

		if r.ctx.Err() != nil {
			return
		}

		execCtx := r.ctx
		var cancel context.CancelFunc

		if r.timeout > 0 {
			execCtx, cancel = context.WithTimeout(r.ctx, r.timeout)
		} else {
			execCtx, cancel = context.WithCancel(r.ctx)
		}

		err := func() (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					err = errors.New("panic recovered")
				}
			}()
			return r.job(execCtx)
		}()

		cancel()

		if err == nil {
			finalErr = nil
			break
		}

		finalErr = err

		if i < r.retry {
			retryTimes++ // ⭐ 这里统计一次重试
			select {
			case <-time.After(r.retryDelay):
			case <-r.ctx.Done():
				return
			}
		}
	}

	atomic.AddInt64(&r.totalRetryCount, retryTimes)
	atomic.StoreInt64(&r.lastRetryCount, retryTimes)
	atomic.AddInt64(&r.runCount, 1)

	if r.onComplete != nil {
		func() {
			defer func() { recover() }()
			r.onComplete(r.id, finalErr)
		}()
	}
}

func (r *Runner) close() {
	if !atomic.CompareAndSwapInt32(&r.closed, 0, 1) {
		return
	}

	r.cancel()

	if r.taskType == TaskCron {
		r.s.cron.Remove(r.entryID)
	}

	if r.onClose != nil {
		func() {
			defer func() { recover() }()
			r.onClose(r.id)
		}()
	}
}

func (s *Scheduler) AddInterval(
	id string,
	interval time.Duration,
	timeout time.Duration,
	retry int,
	retryDelay time.Duration,
	expireAt time.Time, // ⭐ 新增
	job func(context.Context) error,
	onComplete func(id string, err error),
	onClose func(id string),
) (*Runner, error) {

	if atomic.LoadInt32(&s.stopped) == 1 {
		return nil, errors.New("scheduler stopped")
	}

	if interval <= 0 {
		return nil, errors.New("invalid interval")
	}

	if _, exists := s.tasks.Load(id); exists {
		return nil, errors.New("task id already exists")
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := &Runner{
		id:         id,
		taskType:   TaskInterval,
		interval:   interval,
		timeout:    timeout,
		retry:      retry,
		retryDelay: retryDelay,
		jitter:     0.2,
		expireAt:   expireAt,
		job:        job,
		onComplete: onComplete,
		onClose:    onClose,
		ctx:        ctx,
		cancel:     cancel,
		s:          s,
	}

	r.nextRun.Store(r.calcNextIntervalRun())
	s.tasks.Store(id, r)
	go func() {
		for {
			select {
			case <-time.After(time.Until(r.NextRunTime())):
				if !r.expireAt.IsZero() && time.Now().After(r.expireAt) {
					r.s.Remove(r.id)
					return
				}

				if atomic.LoadInt32(&r.closed) == 1 {
					return
				}

				r.execute()
				r.nextRun.Store(r.calcNextIntervalRun())
			case <-r.ctx.Done():
				return
			}
		}
	}()

	return r, nil
}

func (s *Scheduler) AddCron(
	id string,
	expr string,
	timeout time.Duration,
	retry int,
	retryDelay time.Duration,
	expireAt time.Time, // ⭐ 新增
	job func(context.Context) error,
	onComplete func(id string, err error),
	onClose func(id string),
) error {

	if atomic.LoadInt32(&s.stopped) == 1 {
		return errors.New("scheduler stopped")
	}

	if _, exists := s.tasks.Load(id); exists {
		return errors.New("task id already exists")
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := &Runner{
		id:         id,
		taskType:   TaskCron,
		cronExpr:   expr,
		timeout:    timeout,
		retry:      retry,
		retryDelay: retryDelay,
		jitter:     0.2,
		expireAt:   expireAt,
		job:        job,
		onComplete: onComplete,
		onClose:    onClose,
		ctx:        ctx,
		cancel:     cancel,
		s:          s,
	}

	// 先定义执行函数（使用 r.entryID，不用外部变量）
	wrapped := func() {
		if !r.expireAt.IsZero() && time.Now().After(r.expireAt) {
			r.s.Remove(r.id)
			return
		}

		if atomic.LoadInt32(&r.closed) == 1 {
			return
		}

		r.execute()

		entry := s.cron.Entry(r.entryID)
		if !entry.Next.IsZero() {
			r.nextRun.Store(r.calcNextCronRun(entry.Next))
		}
	}

	entryID, err := s.cron.AddFunc(expr, wrapped)
	if err != nil {
		cancel()
		return err
	}

	r.entryID = entryID

	// 初始化 nextRun
	entry := s.cron.Entry(r.entryID)
	if !entry.Next.IsZero() {
		r.nextRun.Store(r.calcNextCronRun(entry.Next))
	}

	s.tasks.Store(id, r)

	return nil
}

func (s *Scheduler) Remove(id string) {
	if v, ok := s.tasks.Load(id); ok {
		r := v.(*Runner)
		r.close()
		s.tasks.Delete(id)
	}
}

func (r *Runner) TotalRetryCount() int64 {
	return atomic.LoadInt64(&r.totalRetryCount)
}

func (r *Runner) LastRetryCount() int64 {
	return atomic.LoadInt64(&r.lastRetryCount)
}

func (s *Scheduler) GetRetryStats(id string) (total int64, last int64, ok bool) {
	v, ok := s.tasks.Load(id)
	if !ok {
		return 0, 0, false
	}
	r := v.(*Runner)
	return r.TotalRetryCount(), r.LastRetryCount(), true
}

func (r *Runner) SetExpireAt(t time.Time) {
	r.expireAt = t
}

func (s *Scheduler) RunNow(id string) bool {
	v, ok := s.tasks.Load(id)
	if !ok {
		return false
	}

	r := v.(*Runner)

	if atomic.LoadInt32(&r.closed) == 1 {
		return false
	}

	go r.execute()

	return true
}
