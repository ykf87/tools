package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type TaskFunc func(context.Context) error
type ErrFun func(error, int32)
type CloseFun func()

type Runner struct {
	id string

	task     TaskFunc
	errFunc  ErrFun
	closeFun CloseFun
	oncedone func(int32, error, time.Time)

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
	stopAt    time.Time // 自动停止时间
	runTimers int
	randesesk float64 // 随机范围,百分比0-1的数字

	mu sync.Mutex
	s  *Scheduler

	firstRun atomic.Bool // 🔥 是否已经执行过

	daily       bool
	dailyHour   int
	dailyMin    int
	dailySec    int
	dailyJitter int
	dailyLoc    *time.Location
}

func newRunner(ctx context.Context, cancel context.CancelFunc, task TaskFunc, s *Scheduler) *Runner {
	return &Runner{
		task:      task,
		ctx:       ctx,
		cancel:    cancel,
		s:         s,
		randesesk: 0.24,
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
	// 已关闭或超时停止
	if r.closed.Load() {
		return
	}

	// 🔥 截止时间判断
	if !r.stopAt.IsZero() && time.Now().After(r.stopAt) {
		r.Stop()
		return
	}

	if r.task == nil || r.ctx.Err() != nil {
		return
	}

	// fmt.Println(r.id, "----")

	r.startAt = time.Now()
	err := r.task(r.ctx)

	var nextTime time.Time
	var needReschedule bool

	if err != nil {
		n := r.tried.Add(1)

		// 未达到最大重试次数 -> retry
		if r.maxTry == 0 || n < int32(r.maxTry) {
			delay := r.retryDelay
			if delay <= 0 {
				delay = 5 * time.Second
			}

			nextTime = time.Now().Add(r.randomizeDelay(delay))
			needReschedule = true

			if r.errFunc != nil {
				r.errFunc(err, n)
			}
		} else {
			// 达到最大重试次数
			tried := r.tried.Load()
			r.tried.Store(0)

			if r.daily {
				nextTime = NextDailyRandomTime(
					time.Now(),
					r.dailyHour,
					r.dailyMin,
					r.dailySec,
					r.dailyJitter,
					r.dailyLoc,
				)
				needReschedule = true
			} else if r.interval > 0 {
				nextTime = time.Now().Add(r.randomizeDelay(r.interval))
				needReschedule = true
			} else {
				r.closed.Store(true)
			}

			if r.oncedone != nil {
				r.oncedone(tried, err, nextTime)
			}
		}

	} else {
		// ========================
		// ✅ 执行成功
		// ========================

		tried := r.tried.Load()
		r.tried.Store(0)
		r.runTimers++

		if r.daily {
			nextTime = NextDailyRandomTime(
				time.Now(),
				r.dailyHour,
				r.dailyMin,
				r.dailySec,
				r.dailyJitter,
				r.dailyLoc,
			)
			needReschedule = true

		} else if r.interval > 0 {
			nextTime = time.Now().Add(r.randomizeDelay(r.interval))
			needReschedule = true
		} else {
			r.closed.Store(true)
		}

		if r.oncedone != nil {
			r.oncedone(tried, nil, nextTime)
		}
	}

	// ========================
	// 🔁 统一调度出口
	// ========================
	if needReschedule && !r.closed.Load() && r.ctx.Err() == nil {
		r.nextRun = nextTime
		r.s.enqueue(r)
		return
	}

	// 真正结束
	if r.closed.CompareAndSwap(false, true) {
		r.endAt = time.Now()
		if r.closeFun != nil {
			r.closeFun()
		}
	}
}

func (r *Runner) randomizeDelay(delay time.Duration) time.Duration {
	if r.randesesk <= 0 {
		return delay
	}

	dlc := float64(delay) * r.randesesk
	offset := (rand.Float64()*2 - 1) * dlc
	return time.Duration(float64(delay) + offset)
}

// 设置下一次的执行时间
func (r *Runner) setNextRunTime(delay time.Duration) {
	if r.randesesk > 0 {
		dlc := float64(delay) * r.randesesk
		offset := (rand.Float64()*2 - 1) * dlc

		delay = time.Duration(float64(delay) + offset)
	}
	r.nextRun = time.Now().Add(delay)
}

/**************** Runner 生命周期 ****************/

func (r *Runner) Stop() {
	fmt.Println("任务执行关闭了--------")
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
		fmt.Println("任务正在执行-----~~~~~~~~")
		return
	}

	// 一次性任务：如果没有 nextRun，默认不调度
	if r.interval > 0 && r.nextRun.IsZero() {
		// r.nextRun = time.Now().Add(r.interval)
		r.setNextRunTime(r.interval)
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

// DailyRandomAt(3, 0, 0, 10, nil)
// 每天 03:00 ±10 分钟
func (r *Runner) DailyRandomAt(
	hour, min, sec int,
	jitterMinutes int,
	loc *time.Location,
) *Runner {
	fmt.Println("执行时间: ", hour, min, sec)
	if loc == nil {
		loc = time.Local
	}

	r.daily = true
	r.dailyHour = hour
	r.dailyMin = min
	r.dailySec = sec
	r.dailyJitter = jitterMinutes
	r.dailyLoc = loc

	r.nextRun = NextDailyRandomTime(
		time.Now(),
		hour, min, sec,
		jitterMinutes,
		loc,
	)
	return r

	// // 包一层 task（只包一次）
	// originTask := r.task
	// r.task = func(ctx context.Context) error {
	// 	err := originTask(ctx)

	// 	// 不管成功失败，都算明天
	// 	next := NextDailyRandomTime(
	// 		time.Now(),
	// 		hour, min, sec,
	// 		jitterMinutes,
	// 		loc,
	// 	)

	// 	r.nextRun = next
	// 	r.s.enqueue(r)

	// 	return err
	// }

	// // 第一次执行时间
	// r.nextRun = NextDailyRandomTime(
	// 	time.Now(),
	// 	hour, min, sec,
	// 	jitterMinutes,
	// 	loc,
	// )

	// return r
}

// 设置最大重试次数
func (r *Runner) SetMaxTry(n int) *Runner {
	if n > 0 {
		r.maxTry = n
	}
	return r
}

// 设置错误回调
func (r *Runner) SetError(fn ErrFun) *Runner {
	r.errFunc = fn
	return r
}

// 设置任务关闭回调
func (r *Runner) SetCloser(fn CloseFun) *Runner {
	r.closeFun = fn
	return r
}

// 设置重试间隔时间
func (r *Runner) SetRetryDelay(d time.Duration) *Runner {
	r.retryDelay = d
	return r
}

// 执行完成一次后的回调
func (r *Runner) SetOnceDone(fn func(int32, error, time.Time)) *Runner {
	r.oncedone = fn
	return r
}

// 获取runner的id
func (r *Runner) GetID() string {
	return r.id
}

// 获取执行器的上下文
func (r *Runner) GetCtx() context.Context {
	return r.ctx
}

// 获取已执行次数
func (r *Runner) GetRunTimes() int {
	return r.runTimers
}

// 获取当次执行时间
func (r *Runner) GetSigleRunTime() float64 {
	tm := time.Now()
	if !r.endAt.IsZero() {
		tm = r.endAt
	}
	cost := tm.Sub(r.startAt)
	return cost.Seconds()
}

// 获取执行器总执行时间
func (r *Runner) GetTotalTime() float64 {
	tm := time.Now()
	if !r.endAt.IsZero() {
		tm = r.endAt
	}
	cost := tm.Sub(r.begin)
	return cost.Seconds()
}

// 获取已重试的次数
func (r *Runner) GetTryTimers() int {
	return int(r.tried.Load())
}

// 获取下次执行时间
func (r *Runner) GetNextRunTime() time.Time {
	return r.nextRun
}

// 设置任务在什么时间停止
func (r *Runner) SetStopAt(t time.Time) *Runner {
	if !t.IsZero() {
		r.stopAt = t
	}
	return r
}

// 获取执行器开始的时间
func (r *Runner) GetStartAt() time.Time {
	return r.startAt
}

// 判断执行请的运行状态
func (r *Runner) IsRuning() bool {
	return !r.closed.Load()
}
