package scheduler

import (
	"context"
	"time"
)

type TaskFunc func(ctx context.Context) error
type TaskStatus int

const (
	TaskWaiting TaskStatus = iota
	TaskRunning
	TaskSuccess
	TaskFailed
)

type Task struct {
	ID string

	Interval time.Duration
	NextRun  time.Time

	Timeout time.Duration

	MaxRetry int
	RetryGap time.Duration

	Run  TaskFunc
	Stop TaskFunc

	Paused bool

	// === 状态字段 ===
	Status     TaskStatus
	LastError  string
	LastRun    time.Time
	FinishTime time.Time

	retryCount int
	index      int
	mutex      chan struct{} // ⭐ 单任务互斥
	Canceled   bool
}
