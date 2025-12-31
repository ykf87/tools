package tasks

import (
	"context"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	db "tools/runtimes/db"
)

var dbs = db.TaskDB

// ========================
// 注入接口定义
// ========================

type TaskLoader func() ([]Task, error)
type TaskExecutor func(ctx context.Context, task *Task, runID int64) error

// ========================
// Task Runtime
// ========================

type TaskRuntime struct {
	Task       *Task
	NextRunAt  int64
	RetryCount int
	Cancel     context.CancelFunc
	Running    atomic.Bool
}

// ========================
// Scheduler
// ========================

type TaskScheduler struct {
	mu        sync.RWMutex
	tasks     map[int64]*TaskRuntime
	tempTasks map[int64]*TaskRuntime // 临时任务
	wake      chan struct{}
	loadFn    TaskLoader
	executeFn TaskExecutor
}

var scheduler *TaskScheduler

func InitScheduler(loader TaskLoader, executor TaskExecutor) {
	scheduler = NewTaskScheduler(
		// Loader 注入
		loader,

		// Executor 注入, 任务具体如何执行在这里
		executor,
	)
	scheduler.Start()
}

// ========================
// 创建调度器（注入）
// ========================

func NewTaskScheduler(loader TaskLoader, executor TaskExecutor) *TaskScheduler {
	return &TaskScheduler{
		tasks:     make(map[int64]*TaskRuntime),
		tempTasks: make(map[int64]*TaskRuntime),
		wake:      make(chan struct{}, 1),
		loadFn:    loader,
		executeFn: executor,
	}
}

// ========================
// 启动
// ========================

func (s *TaskScheduler) Start() {
	s.reload()
	go s.loop()
	go s.hotReload()
}

// ========================
// 调度循环
// ========================
func (s *TaskScheduler) loop() {
	for {
		now := time.Now().Unix()
		var due []*TaskRuntime
		nextWake := now + 3600

		s.mu.RLock()
		for _, rt := range s.tasks {
			if !rt.Running.Load() && rt.NextRunAt <= now {
				due = append(due, rt)
			}
			if rt.NextRunAt > now && rt.NextRunAt < nextWake {
				nextWake = rt.NextRunAt
			}
		}
		s.mu.RUnlock()

		// 优先级排序
		sort.Slice(due, func(i, j int) bool {
			return due[i].Task.Priority > due[j].Task.Priority
		})

		for _, rt := range due {
			s.runTask(rt)
		}

		sleep := time.Duration(nextWake-now) * time.Second
		if sleep <= 0 {
			sleep = time.Second
		}

		select {
		case <-time.After(sleep):
		case <-s.wake:
		}
	}
}

// ========================
// 执行任务
// ========================
func (s *TaskScheduler) runTask(rt *TaskRuntime) {
	// 严格禁止并发
	if !rt.Running.CompareAndSwap(false, true) {
		return
	}

	startAt := time.Now().Unix()

	baseCtx := context.Background()
	var ctx context.Context
	if rt.Task.Timeout > 0 {
		ctx, rt.Cancel = context.WithTimeout(
			baseCtx,
			time.Duration(rt.Task.Timeout)*time.Second,
		)
	} else {
		ctx, rt.Cancel = context.WithCancel(baseCtx)
	}

	go func() {
		defer rt.Running.Store(false)

		// ===== TaskRun 开始 =====
		run := TaskRun{
			TaskID:  rt.Task.ID,
			RunTime: startAt,
		}
		dbs.Create(&run)

		err := s.executeFn(ctx, rt.Task, run.ID)

		doneAt := time.Now().Unix()
		run.DoneTime = doneAt

		if err != nil {
			run.RunStatus = -1
			run.StatusMsg = err.Error()
			rt.RetryCount++
		} else {
			run.RunStatus = 1
			rt.RetryCount = 0
		}
		dbs.Save(&run)

		// ===== 失败重试优先 =====
		if err != nil && rt.RetryCount <= rt.Task.RetryMax {
			// 重试仍然遵循“完成时间基准”
			rt.NextRunAt = doneAt + 3
			return
		}

		// ===== 周期调度（完成时间基准）=====
		if rt.Task.Cycle > 0 {

			// 下一个“理论执行时间”
			next := doneAt + rt.Task.Cycle

			if rt.Task.CatchUp {
				// 补跑模式：如果已经错过多个周期，立即补一个
				now := time.Now().Unix()
				if next <= now {
					rt.NextRunAt = now
				} else {
					rt.NextRunAt = next
				}
			} else {
				// 非补跑：永远只跑一次，不追历史
				rt.NextRunAt = next
			}

		} else {
			// 非周期任务：永不再执行
			rt.NextRunAt = math.MaxInt64
		}
	}()
}

// ========================
// 热更新
// ========================

func (s *TaskScheduler) hotReload() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.reload()
	}
}

func (s *TaskScheduler) reload() {
	list, err := s.loadFn()
	if err != nil {
		log.Println("load task failed:", err)
		return
	}

	now := time.Now().Unix()

	s.mu.Lock()
	defer s.mu.Unlock()

	exists := make(map[int64]struct{})

	for _, t := range list {
		t := t
		exists[t.ID] = struct{}{}

		rt, ok := s.tasks[t.ID]
		if !ok {
			s.tasks[t.ID] = &TaskRuntime{
				Task:      &t,
				NextRunAt: calcNextRun(&t, now),
			}
			continue
		}
		rt.Task = &t
	}

	// 删除任务
	for id, rt := range s.tasks {
		if _, ok := exists[id]; !ok {
			if rt.Cancel != nil {
				rt.Cancel()
			}
			delete(s.tasks, id)
		}
	}

	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func calcNextRun(t *Task, now int64) int64 {
	if t.Starttime > 0 && t.Starttime > now {
		return t.Starttime
	}
	return now
}

// ========================
// 手动停止
// ========================

func (s *TaskScheduler) StopTask(taskID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rt, ok := s.tasks[taskID]; ok {
		if rt.Cancel != nil {
			rt.Cancel()
		}
		delete(s.tasks, taskID)
	}
}

func NotifyTaskChanged(taskID int64) {
	select {
	case scheduler.wake <- struct{}{}:
	default:
	}
}
