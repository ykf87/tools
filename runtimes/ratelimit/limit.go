// 使用方式:
// limiter := ratelimit.New(
//
//	ratelimit.WithLimit(10, 10*time.Second), // 10秒10次
//	ratelimit.WithConcurrency(5),             // 最大5并发
//	ratelimit.WithQueue(1000),                // 队列长度
//
// )
//
//	limiter.Submit(func(ctx context.Context) {
//	    fmt.Println("run task")
//	})
//
// time.Sleep(3 * time.Second)
// limiter.Close()
package ratelimit

import (
	"context"
	"sync"
	"time"
)

type Task func(ctx context.Context)

type Limiter struct {
	ctx    context.Context
	cancel context.CancelFunc

	tokens chan struct{}
	tasks  chan Task

	wg sync.WaitGroup
}

type Option func(*config)

type config struct {
	limit       int
	per         time.Duration
	queue       int
	concurrency int
}

func defaultConfig() config {
	return config{
		limit:       10,
		per:         time.Second,
		queue:       1024,
		concurrency: 1,
	}
}

func WithLimit(n int, per time.Duration) Option {
	return func(c *config) {
		c.limit = n
		c.per = per
	}
}

func WithQueue(n int) Option {
	return func(c *config) {
		c.queue = n
	}
}

func WithConcurrency(n int) Option {
	return func(c *config) {
		c.concurrency = n
	}
}

func New(opts ...Option) *Limiter {

	cfg := defaultConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())

	l := &Limiter{
		ctx:    ctx,
		cancel: cancel,
		tokens: make(chan struct{}, cfg.limit),
		tasks:  make(chan Task, cfg.queue),
	}

	// 初始化 token
	for range cfg.limit {
		l.tokens <- struct{}{}
	}

	l.start(cfg)

	return l
}

func (l *Limiter) start(cfg config) {

	// token refill
	l.wg.Go(func() {
		ticker := time.NewTicker(cfg.per)
		defer ticker.Stop()

		for {
			select {
			case <-l.ctx.Done():
				return

			case <-ticker.C:
				for range cfg.limit {
					select {
					case l.tokens <- struct{}{}:
					default:
					}
				}
			}
		}
	})

	// workers
	for range cfg.concurrency {
		l.wg.Go(func() {

			for {
				select {
				case <-l.ctx.Done():
					return

				case task := <-l.tasks:

					select {
					case <-l.tokens:
						task(l.ctx)

					case <-l.ctx.Done():
						return
					}
				}
			}
		})
	}
}

func (l *Limiter) Submit(task Task) {

	select {
	case l.tasks <- task:

	case <-l.ctx.Done():
	}
}

func (l *Limiter) Close() {
	l.cancel()
	l.wg.Wait()
}
