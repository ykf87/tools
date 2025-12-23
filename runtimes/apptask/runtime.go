package apptask

import (
	"sync"
	"time"
)

type runtime struct {
	opt Options

	mu sync.Mutex

	// taskId -> runtimeTask
	tasks map[int64]*runtimeTask

	// deviceId -> Delivery（WS / API 二选一）
	devices map[string]Delivery
}

type runtimeTask struct {
	task *AppTask

	running bool

	runId     int64
	startedAt int64
	timeoutAt int64

	lastDone int64
}

// =====================
// lifecycle
// =====================

func newRuntime(opt Options) *runtime {
	return &runtime{
		opt:     opt,
		tasks:   make(map[int64]*runtimeTask),
		devices: make(map[string]Delivery),
	}
}

// =====================
// device bind
// =====================

func (r *runtime) bindDevice(deviceId string, d Delivery) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[deviceId] = d
}

func (r *runtime) unbindDevice(deviceId string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.devices, deviceId)
}

// =====================
// task register
// =====================

func (r *runtime) addTask(t *AppTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[t.Id]; ok {
		return ErrTaskExists
	}

	r.tasks[t.Id] = &runtimeTask{
		task: t,
	}
	return nil
}

func (r *runtime) removeTask(taskId int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if rt, ok := r.tasks[taskId]; ok {
		if rt.running {
			return ErrTaskRunning
		}
		delete(r.tasks, taskId)
		return nil
	}
	return ErrTaskNotFound
}

// =====================
// scheduler tick
// =====================

func (r *runtime) tick() {
	now := time.Now().Unix()

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rt := range r.tasks {
		if rt.running {
			// timeout
			if rt.timeoutAt > 0 && now >= rt.timeoutAt {
				go r.finish(rt, 2, "task timeout")
			}
			continue
		}

		t := rt.task

		if t.Starttime > 0 && now < t.Starttime {
			continue
		}
		if t.Endtime > 0 && now > t.Endtime {
			continue
		}

		if t.Cycle > 0 && rt.lastDone > 0 && now-rt.lastDone < t.Cycle {
			continue
		}

		go r.start(rt)
	}
}

// =====================
// start / finish
// =====================

func (r *runtime) start(rt *runtimeTask) {
	r.mu.Lock()
	if rt.running {
		r.mu.Unlock()
		return
	}

	d, ok := r.devices[rt.task.DeviceId]
	if !ok {
		r.mu.Unlock()
		return // device offline
	}

	run, err := r.opt.Store.CreateRun(rt.task)
	if err != nil {
		r.mu.Unlock()
		return
	}

	rt.running = true
	rt.runId = run.RunId
	rt.startedAt = time.Now().Unix()

	if rt.task.Timeout > 0 {
		rt.timeoutAt = rt.startedAt + rt.task.Timeout
	}

	r.mu.Unlock()

	if err := d.Deliver(rt.task); err != nil {
		r.finish(rt, 2, err.Error())
	}
}

func (r *runtime) finish(rt *runtimeTask, status int, msg string) {
	r.mu.Lock()
	if !rt.running {
		r.mu.Unlock()
		return
	}

	runId := rt.runId
	rt.running = false
	rt.runId = 0
	rt.startedAt = 0
	rt.timeoutAt = 0
	rt.lastDone = time.Now().Unix()
	r.mu.Unlock()

	_ = r.opt.Store.FinishRun(runId, status, msg)
}

// =====================
// client report
// =====================

func (r *runtime) report(runId int64, status int, msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rt := range r.tasks {
		if rt.running && rt.runId == runId {
			go r.finish(rt, status, msg)
			return
		}
	}
	// 已被 timeout / force stop，忽略
}

// =====================
// force stop (admin)
// =====================

func (r *runtime) forceStop(taskId int64, operator string) error {
	r.mu.Lock()
	rt, ok := r.tasks[taskId]
	if !ok {
		r.mu.Unlock()
		return ErrTaskNotFound
	}
	if !rt.running {
		r.mu.Unlock()
		return ErrTaskNotRunning
	}

	runId := rt.runId
	r.mu.Unlock()

	// DB 层审计
	if fs, ok := r.opt.Store.(interface {
		ForceStopRun(int64, string) error
	}); ok {
		_ = fs.ForceStopRun(runId, operator)
	} else {
		_ = r.opt.Store.FinishRun(runId, 2, "force stopped by "+operator)
	}

	// 本地释放
	r.mu.Lock()
	rt.running = false
	rt.runId = 0
	rt.startedAt = 0
	rt.timeoutAt = 0
	rt.lastDone = time.Now().Unix()
	r.mu.Unlock()

	return nil
}

// =====================
// api pull mode
// =====================

func (r *runtime) pull(deviceId string) *AppTask {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rt := range r.tasks {
		if rt.running {
			continue
		}
		if rt.task.DeviceId != deviceId {
			continue
		}
		return rt.task
	}
	return nil
}
