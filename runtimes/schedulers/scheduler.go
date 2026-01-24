package schedulers

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Schedulers struct {
	ctx     context.Context    // 调度器上下文
	cancle  context.CancelFunc // 调度器上下文关闭句柄
	mu      sync.Mutex         // 调度器锁
	runner  map[string]*Runner // 被执行的任务
	started atomic.Bool        // 是否启动
	ticker  time.Duration      // 频率
}

type TaskFun func(ctx context.Context) error

type Runner struct {
	mu     sync.Mutex         // 锁
	uuid   string             // 任务唯一编号
	ctx    context.Context    // 调度器上下文
	cancle context.CancelFunc // 调度器上下文关闭句柄

	timer       time.Duration // 任务执行频率
	nextRunTime time.Time     // 下一次执行时间
	task        TaskFun       // 执行任务的方法

	maxTry  int          // 最大重试次数
	tried   atomic.Int32 // 已经重试的次数
	running atomic.Bool  // 是否在运行
	closed  atomic.Bool  // 是否已经被关闭
}

// 总调度器,调度器内部可用有多个任务,这个调度器用于调度内部任务
// 任务通过Add注册
// 调度器必须Start,任务添加前和添加后启动都可以
func New() *Schedulers {
	ctx, cancle := context.WithCancel(context.Background())
	return &Schedulers{
		ctx:    ctx,
		cancle: cancle,
		runner: make(map[string]*Runner),
	}
}

func (s *Schedulers) Start() {
	if s.started.Load() {
		return
	}
	s.mu.Lock()
	s.started.Store(true)
	s.mu.Unlock()

	if s.ticker <= 0 {
		s.ticker = time.Second
	}
	go func() {
		tic := time.NewTicker(s.ticker)
		defer tic.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-tic.C:
				now := time.Now()
				s.mu.Lock()
				for _, v := range s.runner {
					if v.ctx.Err() != nil {
						continue
					}
					if !v.running.Load() && !v.closed.Load() && now.After(v.nextRunTime) {
						go v.Run()
					}
				}
				s.mu.Unlock()
			}
		}
	}()
}

func (s *Schedulers) ReStart() {
	s.Stop()

	ctx, cancle := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancle = cancle

	for _, v := range s.runner {
		ctx, cancle := context.WithCancel(ctx)
		v.cancle = cancle
		v.ctx = ctx
	}
	s.Start()
}

// stop
func (s *Schedulers) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.runner != nil {
		for _, v := range s.runner {
			v.Stop()
		}
	}

	if s.cancle != nil {
		s.cancle()
	}
	s.started.Store(false)
}

func (s *Schedulers) NewRunner(timer time.Duration, task TaskFun, maxtry int) (*Runner, error) {
	uuid := uuid.NewString()

	ctx, cancle := context.WithCancel(s.ctx)

	runn := &Runner{
		uuid:   uuid,
		timer:  timer,
		ctx:    ctx,
		cancle: cancle,
		maxTry: maxtry,
		task:   task,
	}

	s.mu.Lock()
	s.runner[uuid] = runn
	s.mu.Unlock()

	return runn, nil
}

func (r *Runner) Run() error {
	if r.task == nil {
		return fmt.Errorf("执行的方法未传入..., 任务启动失败")
	}
	if !r.running.CompareAndSwap(false, true) {
		return nil
	}
	defer func() {
		r.initNextRun()
		r.running.Store(false)
	}()

	if err := r.task(r.ctx); err != nil {
		if r.tried.Add(1) >= int32(r.maxTry) {
			r.cancle()
			r.closed.Store(true)
			return err
		}
	}

	return nil
}

func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// if !r.running.Load() {
	// 	return
	// }

	r.cancle()
	r.running.Store(false)
	r.closed.Store(true)
}

func (r *Runner) initNextRun() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.timer <= 0 {
		return
	}

	if r.nextRunTime.IsZero() {
		r.nextRunTime = time.Now().Add(r.timer)
	} else {
		r.nextRunTime = r.nextRunTime.Add(r.timer)
	}
}
