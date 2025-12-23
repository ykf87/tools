package apptask

import (
	"sync"
	"time"
)

type runtimeTask struct {
	task     *AppTask
	running  bool
	lastDone int64
}

type runtimeManager struct {
	mu    sync.Mutex
	tasks map[int64]*runtimeTask
	opt   Options
}

func newRuntime(opt Options) *runtimeManager {
	return &runtimeManager{
		tasks: make(map[int64]*runtimeTask),
		opt:   opt,
	}
}

func (r *runtimeManager) add(task *AppTask) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.Id] = &runtimeTask{task: task}
}

func (r *runtimeManager) tick() {
	now := time.Now().Unix()

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rt := range r.tasks {
		t := rt.task

		if rt.running {
			continue
		}

		if t.Starttime > 0 && now < t.Starttime {
			continue
		}

		if t.Endtime > 0 && now > t.Endtime {
			continue
		}

		if t.Cycle > 0 && rt.lastDone > 0 {
			if now-rt.lastDone < t.Cycle {
				continue
			}
		}

		// 串行执行（单任务不会并发）
		rt.running = true
		go r.execute(rt)
	}
}

func (r *runtimeManager) execute(rt *runtimeTask) {
	run, _ := r.opt.Store.CreateRun(rt.task)

	err := r.opt.Executor.Execute(rt.task)

	status := 1
	msg := "ok"
	if err != nil {
		status = 2
		msg = err.Error()
	}

	r.opt.Store.FinishRun(run.RunId, status, msg)

	r.mu.Lock()
	rt.running = false
	rt.lastDone = time.Now().Unix()
	r.mu.Unlock()
}
