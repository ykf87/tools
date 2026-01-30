package scheduler

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type TaskFunc func(ctx context.Context) error
type ErrFun func(err error)
type CloseFun func()

type Runner struct {
	id string

	task     TaskFunc
	errFunc  ErrFun
	closeFun CloseFun

	interval   time.Duration // 0 = 只执行一次
	nextRun    time.Time
	retryDelay time.Duration

	maxTry int
	tried  atomic.Int32

	running atomic.Bool
	closed  atomic.Bool

	ctx    context.Context
	cancel context.CancelFunc

	begin     time.Time // 任务开始时间
	startAt   time.Time // 单次任务开始执行时间
	endAt     time.Time // 任务结束时间
	runTimers int

	mu sync.Mutex
	s  *Scheduler

	firstRun atomic.Bool // 🔥 是否已经执行过
}

func newRunner(ctx context.Context, cancel context.CancelFunc, task TaskFunc, s *Scheduler) *Runner {
	return &Runner{
		task:   task,
		ctx:    ctx,
		cancel: cancel,
		s:      s,
	}
}

/**************** 执行核心 ****************/

func (r *Runner) execute() {
	if !r.running.CompareAndSwap(false, true) {
		return
	}
	defer r.running.Store(false)
	// 🔥 标记：已经至少执行过一次
	r.firstRun.Store(true)

	if r.task == nil || r.ctx.Err() != nil {
		return
	}

	// fmt.Println(r.id, "----")

	r.startAt = time.Now()
	if err := r.task(r.ctx); err != nil {
		n := r.tried.Add(1)

		if n >= int32(r.maxTry) {
			if r.errFunc != nil {
				r.errFunc(err)
			}
			r.Stop()
			return
		}

		// 🔥 失败重试调度（而不是等 interval）
		delay := r.retryDelay
		if delay <= 0 {
			delay = time.Millisecond // 防止自旋
		}

		r.nextRun = time.Now().Add(delay)
		r.s.enqueue(r)
		return
	}

	// 成功
	r.tried.Store(0)
	r.runTimers++

	// 只有成功，才进入周期调度
	if r.interval > 0 && !r.closed.Load() {
		r.nextRun = time.Now().Add(r.interval)
		r.s.enqueue(r)
		return
	}

	r.closed.Store(true)
	if r.closeFun != nil {
		r.closeFun()
	}
}

/**************** Runner 生命周期 ****************/

func (r *Runner) Stop() {
	if r.closed.CompareAndSwap(false, true) {
		r.endAt = time.Now()
		r.cancel()
		if r.closeFun != nil {
			r.closeFun()
		}
	}
}

/**************** 对外 API（重点） ****************/

// Run：加入调度器，但不立即执行
func (r *Runner) Run() {
	if r.closed.Load() {
		return
	}

	// 一次性任务：如果没有 nextRun，默认不调度
	if r.interval > 0 && r.nextRun.IsZero() {
		r.nextRun = time.Now().Add(r.interval)
	}

	if !r.nextRun.IsZero() {
		r.s.enqueue(r)
	}
}

// RunNow：立即执行一次（仅一次）
func (r *Runner) RunNow() {
	if r.closed.Load() {
		return
	}
	r.nextRun = time.Now()
	r.s.enqueue(r)
}

func (r *Runner) Every(d time.Duration) *Runner {
	if d > 0 {
		r.interval = d
	}
	return r
}

func (r *Runner) SetMaxTry(n int) *Runner {
	if n > 0 {
		r.maxTry = n
	}
	return r
}

func (r *Runner) SetError(fn ErrFun) *Runner {
	r.errFunc = fn
	return r
}

func (r *Runner) SetCloser(fn CloseFun) *Runner {
	r.closeFun = fn
	return r
}

func (r *Runner) SetRetryDelay(d time.Duration) *Runner {
	r.retryDelay = d
	return r
}

func (r *Runner) GetID() string {
	return r.id
}

func (r *Runner) GetCtx() context.Context {
	return r.ctx
}

func (r *Runner) GetRunTimes() int {
	return r.runTimers
}

func (r *Runner) GetSigleRunTime() float64 {
	tm := time.Now()
	if !r.endAt.IsZero() {
		tm = r.endAt
	}
	cost := tm.Sub(r.startAt)
	return cost.Seconds()
}

func (r *Runner) GetTotalTime() float64 {
	tm := time.Now()
	if !r.endAt.IsZero() {
		tm = r.endAt
	}
	cost := tm.Sub(r.begin)
	return cost.Seconds()
}

func (r *Runner) GetTryTimers() int {
	return int(r.tried.Load())
}
