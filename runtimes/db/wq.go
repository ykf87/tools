package db

import (
	"context"
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"
)

/*
设计目标：

- 每个 SQLite 文件一个 SQLiteWriter
- 所有写操作必须通过 Write 提交
- 内部单 goroutine 顺序执行
- 支持优雅关闭 & 等待队列清空
- 不影响读（读直接用 DB()）
*/

var (
	ErrWriterClosed = errors.New("sqlite writer closed")
	ErrQueueFull    = errors.New("sqlite writer queue full")
)

/***************
 * 对外接口定义
 ***************/

type WriteExecutor interface {
	Write(fn func(tx *gorm.DB) error) error
	WriteCtx(ctx context.Context, fn func(tx *gorm.DB) error) error
	DB() *gorm.DB
	Close() error
	Wait()
}

/***************
 * 写任务定义
 ***************/

type writeTask struct {
	ctx  context.Context
	fn   func(tx *gorm.DB) error
	done chan error
}

/***************
 * SQLiteWriter
 ***************/

type SQLiteWriter struct {
	db *gorm.DB

	queue chan *writeTask

	closeOnce sync.Once
	closed    chan struct{}
	wg        sync.WaitGroup
}

/***************
 * 构造函数
 ***************/

// queueSize：建议 500~5000，取决于你瞬时写峰值
func NewSQLiteWriter(db *gorm.DB, queueSize int) *SQLiteWriter {
	if queueSize <= 0 {
		queueSize = 1000
	}

	w := &SQLiteWriter{
		db:     db,
		queue:  make(chan *writeTask, queueSize),
		closed: make(chan struct{}),
	}

	w.wg.Add(1)
	go w.loop()

	return w
}

/***************
 * 主循环
 ***************/

func (w *SQLiteWriter) loop() {
	defer w.wg.Done()

	for task := range w.queue {
		// context 已取消，直接返回
		select {
		case <-task.ctx.Done():
			task.done <- task.ctx.Err()
			close(task.done)
			continue
		default:
		}

		err := w.db.Transaction(func(tx *gorm.DB) error {
			return task.fn(tx)
		})

		task.done <- err
		close(task.done)
	}
}

/***************
 * 写入接口
 ***************/

func (w *SQLiteWriter) Write(fn func(tx *gorm.DB) error) error {
	return w.WriteCtx(context.Background(), fn)
}

func (w *SQLiteWriter) WriteCtx(
	ctx context.Context,
	fn func(tx *gorm.DB) error,
) error {
	select {
	case <-w.closed:
		return ErrWriterClosed
	default:
	}

	task := &writeTask{
		ctx:  ctx,
		fn:   fn,
		done: make(chan error, 1),
	}

	select {
	case w.queue <- task:
		return <-task.done
	default:
		return ErrQueueFull
	}
}

/***************
 * 读接口
 ***************/

func (w *SQLiteWriter) DB() *gorm.DB {
	return w.db
}

/***************
 * 关闭与等待
 ***************/

// Close：
// - 禁止新写入
// - 处理完已入队任务
func (w *SQLiteWriter) Close() error {
	w.closeOnce.Do(func() {
		close(w.closed)
		close(w.queue)
	})
	return nil
}

// Wait：等待写队列完全退出
func (w *SQLiteWriter) Wait() {
	w.wg.Wait()
}

/***************
 * 可选：健康检查
 ***************/

func (w *SQLiteWriter) Ping(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return w.WriteCtx(ctx, func(tx *gorm.DB) error {
		return nil
	})
}
