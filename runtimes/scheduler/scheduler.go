package scheduler

import (
	"container/heap"
	"context"
	"math/rand"
	"sync"
	"time"
	"tools/runtimes/mainsignal"

	"github.com/google/uuid"
)

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu   sync.Mutex
	pq   runnerHeap
	wake chan struct{}

	sem    chan struct{} // 并发数量控制
	jitter time.Duration // 多任务之间随机休眠多久后启动,避免多个任务同时启动
}

func New(ctx context.Context) *Scheduler {
	return NewWithLimit(ctx, 50)
}

// 执行并发现在的创建
func NewWithLimit(parentCtx context.Context, limit int) *Scheduler {
	if parentCtx == nil {
		parentCtx = mainsignal.MainCtx
	}
	ctx, cancel := context.WithCancel(parentCtx)
	s := &Scheduler{
		ctx:    ctx,
		cancel: cancel,
		wake:   make(chan struct{}, 1),
		sem:    make(chan struct{}, limit), // 🔥 最大并发数
	}
	heap.Init(&s.pq)
	go s.loop()
	return s
}

// 多任务之间随机休眠多久后启动,避免多个任务同时启动
func (s *Scheduler) SetJitter(d time.Duration) {
	if d > 0 {
		s.jitter = d
	}
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.signal()
}

// 每天固定时间执行
// jitterMinutes 为随机数,避免每天都固定在某个时间点,单位是分钟
// func (s *Scheduler) DailyRandomAt(
// 	hour, min, sec int,
// 	jitterMinutes int,
// 	task TaskFunc,
// ) *Runner {

// 	loc := time.Local
// 	var r *Runner
// 	r = s.NewRunner(func(ctx context.Context) error {
// 		err := task(ctx)

// 		// 🔥 不管成功失败，都调度明天
// 		next := NextDailyRandomTime(
// 			time.Now(),
// 			hour, min, sec,
// 			jitterMinutes,
// 			loc,
// 		)

// 		r.nextRun = next
// 		r.s.enqueue(r)

// 		return err
// 	}, 0, nil)

// 	// 第一次执行时间
// 	r.nextRun = NextDailyRandomTime(
// 		time.Now(),
// 		hour, min, sec,
// 		jitterMinutes,
// 		loc,
// 	)

// 	// r.Run()
// 	return r
// }

func NextDailyRandomTime(
	now time.Time,
	hour, min, sec int,
	jitterMin int,
	loc *time.Location,
) time.Time {
	if loc == nil {
		loc = time.Local
	}
	n := now.In(loc)

	// 今天的 base 时间
	base := time.Date(
		n.Year(), n.Month(), n.Day(),
		hour, min, sec, 0,
		loc,
	)

	// 如果已经 >= 今天的 base，直接用明天
	if !n.Before(base) {
		base = base.Add(24 * time.Hour)
	}

	// jitter 计算（不允许跨天）
	var offset time.Duration
	if jitterMin > 0 {
		j := time.Duration(jitterMin) * time.Minute

		dayStart := time.Date(
			base.Year(), base.Month(), base.Day(),
			0, 0, 0, 0,
			loc,
		)
		dayEnd := dayStart.Add(24 * time.Hour)

		minOffset := -j
		maxOffset := j

		if base.Add(minOffset).Before(dayStart) {
			minOffset = dayStart.Sub(base)
		}
		if base.Add(maxOffset).After(dayEnd) {
			maxOffset = dayEnd.Sub(base)
		}

		if maxOffset > minOffset {
			offset = minOffset + time.Duration(
				rand.Int63n(int64(maxOffset-minOffset)),
			)
		}
	}

	return base.Add(offset)
}

func (s *Scheduler) NewRunner(task TaskFunc, timeout time.Duration, pctx context.Context) *Runner {
	if pctx == nil {
		pctx = s.ctx
	}
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(pctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(pctx)
	}

	r := newRunner(ctx, cancel, task, s)
	r.begin = time.Now()
	r.id = uuid.NewString()
	return r
}

func (s *Scheduler) enqueue(r *Runner) {
	if r.closed.Load() {
		return
	}
	if !r.stopAt.IsZero() && time.Now().After(r.stopAt) {
		r.Stop()
		return
	}

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

		select {
		case s.sem <- struct{}{}:
		case <-s.ctx.Done():
			return
		}

		go func() {
			defer func() {
				<-s.sem // 🔥 执行完释放
			}()

			// 🔥 随机启动抖动
			// 🔥 只有第一次执行才 jitter
			if s.jitter > 0 && !r.firstRun.Load() {
				d := time.Duration(rand.Int63n(int64(s.jitter)))
				timer := time.NewTimer(d)
				select {
				case <-timer.C:
				case <-s.ctx.Done():
					timer.Stop()
					return
				case <-r.ctx.Done():
					timer.Stop()
					return
				}
			}

			r.execute()
		}()
	}
}
