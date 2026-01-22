package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/jses"
	"tools/runtimes/scheduler"

	"github.com/chromedp/cdproto/runtime"
	"github.com/tidwall/gjson"
)

var tmpIndex int64
var dbTasks sync.Map

// 仅针对浏览器可用设置临时的设备
func TempTask(adminID int64) *Task {
	tmpIndex--
	return &Task{
		ID:     tmpIndex,
		Status: 1,
	}
}

func listen() {
	for {
		dbTasks.Range(func(k, v any) bool {
			if tsk, ok := v.(*Task); ok {
				tsk.watching()
			}
			return true
		})
		time.Sleep(time.Second * 30)
	}
}

func (this *Task) watching() error {
	if this.ID == 0 {
		return fmt.Errorf("任务未设置")
	}

	now := time.Now().Unix()
	if this.Starttime > 0 && this.Starttime > now {
		return nil
	}

	if this.Endtime > 0 && this.Endtime < now {
		return nil
	}
	this.Run()

	return nil
}

// 将任务添加到调度器并启动
// func (this *Task) push() error {
// 	this.mu.Lock()
// 	defer this.mu.Unlock()

// 	if this.ID == 0 {
// 		return fmt.Errorf("任务未设置")
// 	}
// 	if this.Status == 0 {
// 		return fmt.Errorf("任务未启动")
// 	}
// 	if this.runner == nil {
// 		return fmt.Errorf("未设置执行方法")
// 	}

// 	taskID := this.taskID()

// 	if Seched.Exists(taskID) {
// 		return nil
// 	}

// 	go func() {
// 		time.Sleep(time.Second * 14)
// 		Seched.Remove(this.taskID())
// 	}()

// 	return Seched.Add(&scheduler.Task{
// 		ID:       taskID,
// 		Interval: time.Duration(this.Cycle) * time.Second,
// 		Run:      this.runner,
// 		Stop:     this.stop,
// 		MaxRetry: this.RetryMax,
// 		Timeout:  time.Duration(this.Timeout) * time.Second,
// 	})
// }

func (this *Task) Run() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.ID == 0 {
		return fmt.Errorf("任务未设置")
	}
	if this.Status == 0 {
		return fmt.Errorf("任务未启动")
	}
	if this.Callback == nil {
		return fmt.Errorf("未设置执行方法")
	}

	taskID := GenTaskID(this.ID)
	if Seched.Exists(taskID) {
		return nil
	}

	return Seched.Add(&scheduler.Task{
		ID:       taskID,
		Interval: time.Duration(this.Cycle) * time.Second,
		Run:      this.runner,
		Stop:     this.stop,
		MaxRetry: this.RetryMax,
		Timeout:  time.Duration(this.Timeout) * time.Second,
	})
}

func (this *Task) runner(ctx context.Context) error {
	if this.ScriptStr == "" && this.Script > 0 {
		js := jses.GetJsById(this.Script)
		if js != nil && js.ID > 0 {
			params := this.GetParams()
			mp := make(map[string]any)
			for _, v := range params {
				mp[v.CodeName] = v.Value
			}
			this.ScriptStr = js.GetContent(mp)
		}
	}

	var taskDevices []*TaskClients
	if this.ID > 0 {
		dbs.Model(&TaskClients{}).Where("task_id = ?", this.ID).Find(&taskDevices)
	}
	if len(taskDevices) < 1 {
		// taskDevices
	}
	return nil //fmt.Errorf("error")
}

func RemoveTask(id int64) {
	Seched.Remove(GenTaskID(id))
}

func (this *Task) stop(ctx context.Context) error {
	this.isRuning = false
	if tt, ok := dbTasks.Load(this.ID); ok {
		if tsk, ok := tt.(*Task); ok {
			tsk.Status = 0
			tsk.Save(nil)
		}
		dbTasks.Delete(this.ID)
	}
	return nil
}

func GenTaskID(id int64) string {
	return fmt.Sprintf("task-%d", id)
}

// 执行浏览器的任务
func (this *Task) runBrowser(browserID int64) error {
	bbs := bs.NewManager("")
	brows, _ := bbs.New(browserID, bs.Options{
		Url:      this.DefUrl,
		JsStr:    this.ScriptStr,
		Headless: this.Headless == 1,
		Timeout:  time.Duration(time.Second * time.Duration(this.Timeout)),
	})

	if this.OnClose != nil {
		brows.OnClosed(func() {
			this.OnClose()
		})
	}

	if this.Callback != nil {
		brows.OnConsole(func(args []*runtime.RemoteObject) {
			for _, arg := range args {
				if arg.Value != nil {
					gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
					if gs.Get("type").String() == "kaka" {
						this.Callback(gs.Get("data").String())
					}
				}
			}
		})
	}

	if this.OnUrlchange != nil {
		brows.OnURLChange(func(url string) {
			this.OnUrlchange(url)
		})
	}

	brows.OpenBrowser()
	time.Sleep(time.Second * 1)
	go brows.RunJs(this.ScriptStr)
	return nil
}
