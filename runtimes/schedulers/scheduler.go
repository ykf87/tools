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
	ctx    context.Context    // 调度器上下文
	cancle context.CancelFunc // 调度器上下文关闭句柄
	mu     sync.Mutex         // 调度器锁
	runner map[string]*Runner // 被执行的任务
}

type TaskFun func(ctx context.Context) error

type Runner struct {
	uuid    string             // 任务唯一编号
	ctx     context.Context    // 调度器上下文
	cancle  context.CancelFunc // 调度器上下文关闭句柄
	mu      sync.Mutex         // 锁
	timer   time.Duration      // 任务执行频率
	started bool               // 是否启动
	task    TaskFun            // 执行任务的方法
	maxTry  int                // 最大重试次数
	tried   atomic.Int32       // 已经重试的次数
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

func (s *Schedulers) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancle != nil {
		return fmt.Errorf("任务已启动,请使用ReStart重启")
	}

	for _, v := range s.runner {
		go v.Run()
	}
	return nil
}

func (s *Schedulers) ReStart() {
	s.Stop()
	_ = s.Start()
}

// stop后再要使用Schedulers,必须重新new
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
	s.ctx = nil
	s.cancle = nil
}

func (s *Schedulers) Add(runner *Runner) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if runner.
}

func (s *Schedulers) NewRunner(timer time.Duration, task TaskFun, maxtry int) (*Runner, error) {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	ctx, cancle := context.WithCancel(s.ctx)

	runn := &Runner{
		uuid:   uuid.String(),
		timer:  timer,
		ctx:    ctx,
		cancle: cancle,
		maxTry: maxtry,
	}

	s.runner[uuid.String()] = runn

	return runn, nil
}

func (r *Runner) Run() error {
	if r.task == nil {
		return fmt.Errorf("执行的方法未传入..., 任务启动失败")
	}
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return nil
	}
	r.started = true
	r.mu.Unlock()

	if r.timer == 0 {
		go r.task(r.ctx)
	} else {
		go func() {
			ticker := time.NewTicker(r.timer)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := r.task(r.ctx); err != nil {
						if int(r.tried.Load()) >= r.maxTry {
							r.cancle()
							return
						}
						r.tried.Add(1)
					}
				case <-r.ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}

func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.started {
		return
	}

	r.cancle()
}
