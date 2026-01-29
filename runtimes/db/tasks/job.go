package tasks

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/tasks/runhttp"
	"tools/runtimes/db/tasks/runphone"
	"tools/runtimes/db/tasks/runweb"
	"tools/runtimes/mainsignal"
	"tools/runtimes/scheduler"
)

var WatchingTasks sync.Map

type RuningTask struct {
	ID       int64
	ErrMsg   string                 // 任务执行错误消息
	mu       sync.Mutex             // 锁
	Callback func(string) error     // 任务执行结果回调
	OnError  func(error)            // 任务错误结果回调
	OnClose  func()                 // 浏览器关闭回调
	OnChange func(string) error     // 当浏览器地址改变回调
	slots    chan struct{}          // 启动的协程
	runners  map[string]*RunnerData // 任务中具体执行的设备
	isRun    atomic.Bool            // 是否在执行中
	ctx      context.Context
	cancle   context.CancelFunc
}

type RunnerData struct {
	r Runner
	s *scheduler.Runner
}

type Runner interface {
	Start(context.Context) error
	OnError(error)
	OnClose()
	OnChange(string)
	SetRunner(*scheduler.Runner)
}

// 获取任务
func GetRunTask(id int64) (*RuningTask, error) {
	if bt, ok := WatchingTasks.Load(id); ok {
		if btt, ok := bt.(*RuningTask); ok {
			return btt, nil
		}
	}
	return nil, errors.New("not found")
}

// 添加任务
func Start(
	t *Task,
	Callback func(string) error,
	OnClose func(),
	OnChange func(string, *bs.Browser) error,
	OnError func(error, *bs.Browser),
) *RuningTask {
	if rt, err := GetRunTask(t.ID); err == nil {
		return rt
	}

	if t.SeNum < 1 {
		t.SeNum = 2
	}
	rt := &RuningTask{
		ID:      t.ID,
		runners: make(map[string]*RunnerData),
		slots:   make(chan struct{}, t.SeNum),
	}
	rt.ctx, rt.cancle = context.WithCancel(mainsignal.MainCtx)

	go func() {
		for _, v := range t.Clients {
			var runner Runner
			switch t.Tp {
			case 0:
				runner = runweb.New(rt.release, &runweb.Option{
					Headless: !(t.Headless == 1),
					Js:       t.GetRunJscript(),
					Url:      t.DefUrl,
					ID:       v.DeviceID,
					Timeout:  time.Duration(t.Timeout) * time.Second,
					OnError:  OnError,
					OnClose:  OnClose,
					OnChange: OnChange,
					Callback: Callback,
					Ctx:      rt.ctx,
				})
			case 1:
				runner = runphone.New(rt.release)
			case 2:
				runner = runhttp.New(rt.release)
			}

			sr := Seched.
				NewRunner(runner.Start, time.Duration(t.Timeout)*time.Second, rt.ctx).
				Every(time.Duration(t.Cycle) * time.Second).
				SetCloser(runner.OnClose).
				SetError(runner.OnError).
				SetMaxTry(t.RetryMax).
				SetRetryDelay(time.Second * 10)
			runner.SetRunner(sr)
			rt.runners[v.GetName()] = &RunnerData{
				r: runner,
				s: sr,
			}

			rt.acquire()
			sr.RunNow()
		}
	}()

	WatchingTasks.Store(t.ID, rt)
	WatchingTasks.Range(func(k, v any) bool {
		fmt.Println(k, "--r")
		return true
	})
	return rt
}

func (rt *RuningTask) acquire() {
	rt.slots <- struct{}{}
}
func (t *RuningTask) release() {
	select {
	case <-t.slots:
	default:
		// 理论上不应该发生，防御式
	}
}

func (t *Task) Listen() {
	if _, err := GetRunTask(t.ID); err == nil {
		return
	}
	now := time.Now().Unix()
	if t.Endtime > 0 && t.Endtime < now {
		t.Status = 0
		t.Save(nil)
		return
	}

	if t.Starttime > 0 && t.Starttime > now {
		delay := time.Duration(t.Starttime-now) * time.Second
		time.AfterFunc(delay, func() {
			t.Clients = t.GetClients()
			go t.Start()
		})
	} else { // 未设置开始时间
		t.Clients = t.GetClients()
		go t.Start()
	}
	if t.Endtime > 0 && t.Endtime > now {
		delay := time.Duration(t.Endtime-now) * time.Second
		time.AfterFunc(delay, func() {
			t.Stop()
		})
	}
	// WatchingTasks.Store(t.ID, t)
	// return
}

// 如果没有设置执行设备,则默认使用golang内置的http发起相应的请求
func (t *Task) Start() *RuningTask {
	return Start(t, nil, nil, nil, nil)
}

func (t *Task) Stop() {
	rt, err := GetRunTask(t.ID)
	WatchingTasks.Range(func(k, v any) bool {
		fmt.Println(k, "--r")
		return true
	})
	if err != nil {
		return
	}
	for _, v := range rt.runners {
		v.s.Stop()
	}
	rt.cancle()
	WatchingTasks.Delete(t.ID)
}
