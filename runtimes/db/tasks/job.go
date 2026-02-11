package tasks

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/task"
	"tools/runtimes/db/tasks/runhttp"
	"tools/runtimes/db/tasks/runphone"
	"tools/runtimes/db/tasks/runweb"
	"tools/runtimes/funcs"
	"tools/runtimes/mainsignal"
	"tools/runtimes/proxy"
	"tools/runtimes/runner"
	"tools/runtimes/scheduler"
)

var WatchingTasks sync.Map
var taskTickerStart sync.Map

type RuningTask struct {
	ID      int64                  `json:"id"`       // 任务的编号
	AdminID int64                  `json:"admin_id"` // 管理员id
	UUID    string                 `json:"uuid"`     // 任务唯一编号
	Title   string                 `json:"title"`    // 任务名称
	Tags    []string               `json:"tags"`     // 任务标签
	ErrMsg  string                 `json:"err_msg"`  // 任务执行错误消息
	Msg     chan string            `json:"-"`        // 任务执行实时消息
	StartAt int64                  `json:"start_at"` // 开启时间
	EndAt   int64                  `json:"end_at"`   // 结束时间
	Total   int                    `json:"total"`    // 执行总数
	mu      sync.Mutex             // 锁
	runners map[string]*RunnerData // 任务中具体执行的设备
	sec     *scheduler.Scheduler   // 调度器
	isRun   atomic.Bool            // 是否在执行中
	ctx     context.Context
	cancle  context.CancelFunc
}

type RunnerData struct {
	r     Runner
	s     *scheduler.Runner
	Title string `json:"title"`
	UUID  string `json:"uuid"`
}

type Runner interface {
	Start(context.Context) error
	OnError(error)
	OnClose()
	OnChange(string)
	SetRunner(*scheduler.Runner)
	Close() // 关闭本次启动的资源
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
	Callback func(string, string) error,
	OnClose func(),
	OnChange func(string, *bs.Browser) error,
	OnError func(error, *bs.Browser),
) *RuningTask {
	if rt, err := GetRunTask(t.ID); err == nil {
		return rt
	}
	if len(t.Clients) < 1 {
		t.Clients = t.GetClients()
	}

	if t.SeNum < 1 {
		t.SeNum = 2
	}
	rt := &RuningTask{
		ID:      t.ID,
		runners: make(map[string]*RunnerData),
		Msg:     make(chan string),
		Title:   t.Title,
		Tags:    t.Tags,
		UUID:    funcs.RoundmUuid(),
		AdminID: t.AdminId,
		Total:   len(t.Clients),
		StartAt: time.Now().Unix(),
	}
	rt.ctx, rt.cancle = context.WithCancel(mainsignal.MainCtx)

	if t.SeNum > 0 {
		rt.sec = scheduler.NewWithLimit(rt.ctx, t.SeNum)
	} else {
		rt.sec = scheduler.New(rt.ctx)
	}
	if t.CycleDelay > 0 {
		rt.sec.SetJitter(time.Second * time.Duration(t.CycleDelay))
	}
	go func() {
		for _, v := range t.Clients {
			tcUUID := v.GetUUID()
			var runner Runner
			switch v.DeviceType {
			case 0:
				runner = runweb.New(nil, &runweb.Option{
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
					UUID:     tcUUID,
				})
			case 1:
				runner = runphone.New(nil)
			case 2:
				runner = runhttp.New(nil)
			}
			sr := rt.sec.
				NewRunner(runner.Start, time.Duration(t.Timeout)*time.Second, nil).
				SetCloser(runner.OnClose).
				// SetError(runner.OnError).
				SetMaxTry(t.RetryMax).
				SetRetryDelay(time.Second * 10)
			if t.Cycle > 0 {
				sr.Every(time.Duration(t.Cycle) * time.Minute)
			}
			runner.SetRunner(sr)
			if t.Endtime > 0 {
				sr.SetStopAt(time.Unix(t.Endtime, 0))
			}
			rt.runners[tcUUID] = &RunnerData{
				r:     runner,
				s:     sr,
				Title: v.GetName(),
				UUID:  tcUUID,
			}

			sr.RunNow()
		}
	}()
	WatchingTasks.Store(t.ID, rt)

	// rt.Sent("")
	// 发送任务的通知
	// bt, _ := config.Json.Marshal(map[string]any{
	// 	"type": "task",
	// 	"data": rt.genRunnerMsg(),
	// })
	// ws.SentMsg(t.AdminId, bt)
	return rt
}

// 重启任务
func (t *Task) ReStart() *RuningTask {
	if rta, ok := WatchingTasks.Load(t.ID); ok {
		if rt, ok := rta.(*RuningTask); ok {
			rt.Stop()
		}
	}
	return t.Start()
}

// 如果没有设置执行设备,则默认使用golang内置的http发起相应的请求
func (t *Task) Start() *RuningTask {
	if t.Endtime > 0 && time.Now().Unix() >= t.Endtime {
		return nil
	}
	tl, err := task.NewTask("task", t.ID, t.Title, t.SeNum, false)
	if err != nil {
		fmt.Println(err, "----- 启动失败")
		return nil
	}
	for _, v := range t.GetClients() {
		tr, err := tl.AddChild(fmt.Sprintf("task-%d-%d", t.ID, v.DeviceID), fmt.Sprintf("设备: %d", v.DeviceID), time.Duration(t.Timeout)*time.Second)
		if err != nil {
			return nil
		}
		tr.StartInterval(t.Cycle*60, func(tr *task.TaskRun) error {
			var opt any
			switch v.DeviceType {
			case 0:
				var bid int64
				var pc *proxy.ProxyConfig
				var w int
				var h int
				var lang string
				var timezone string
				hdl := true
				if bss, err := browserdb.GetBrowserById(v.DeviceID); err == nil {
					bid = bss.Id
					if pcc, err := bss.GenProxyConfig(); err == nil {
						pc = pcc
					}
					w = bss.Width
					h = bss.Height
					lang = bss.Lang
					timezone = bss.Timezone
				}
				if t.Headless == 1 {
					hdl = false
				}
				opt = runner.GenWebOpt(tr.GetCtx(), bid, hdl, t.DefUrl, t.GetRunJscript(), pc, time.Duration(t.Timeout)*time.Second, w, h, lang, timezone)
			case 1:
			case 2:
			}

			rr, err := runner.GetRunner(v.DeviceType, opt)
			if err != nil {
				return nil
			}

			return rr.Start(time.Duration(t.Timeout)*time.Second, func(str string) error {
				fmt.Println(str, "======----")
				return fmt.Errorf("未找到数据----")
			})
		})
		tr.StopAt(time.Unix(t.Endtime, 0))
	}
	return nil
	// now := time.Now().Unix()
	// if t.Endtime > 0 && t.Endtime < now {
	// 	return nil
	// }
	// if t.Starttime > 0 && t.Starttime > now {
	// 	delay := time.Duration(t.Starttime-now) * time.Second
	// 	taskTickerStart.Store(t.ID, true)
	// 	time.AfterFunc(delay, func() {
	// 		t.Start()
	// 	})
	// 	return nil
	// }

	// return Start(t, func(data string, uuid string) error {
	// 	gs := gjson.Parse(data)
	// 	switch gs.Get("type").String() {
	// 	case "done": //任务完成
	// 		fmt.Println("任务完成!")
	// 	case "notify": //任务通知
	// 		fmt.Println("任务通知!")
	// 	case "error": //任务失败
	// 		fmt.Println("任务失败!")
	// 		if wt, ok := WatchingTasks.Load(t.ID); ok {
	// 			if rt, ok := wt.(*RuningTask); ok {
	// 				rt.Stop()
	// 			}
	// 		}
	// 	}
	// 	return nil
	// }, func() { // close
	// 	fmt.Println("任务关闭")
	// }, nil, nil)
}

func (t *Task) Stop() {
	task.Stop("task", t.ID)
	// taskTickerStart.Delete(t.ID)
	// rt, err := GetRunTask(t.ID)
	// if err != nil {
	// 	return
	// }
	// rt.Stop()
}

func (rt *RuningTask) Stop() {
	rt.cancle()
	WatchingTasks.Delete(rt.ID)
}

func Flush() {
	WatchingTasks.Range(func(k, v any) bool {
		if tk, ok := v.(*RuningTask); ok {
			tk.Stop()
		}
		return true
	})
}
