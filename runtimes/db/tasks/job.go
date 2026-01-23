package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/jses"
	"tools/runtimes/funcs"
	"tools/runtimes/scheduler"

	"github.com/chromedp/cdproto/runtime"
	"github.com/tidwall/gjson"
)

var tmpIndex int64
var dbTasks sync.Map
var bbs *bs.Manager

func init() {
	bbs = bs.NewManager("")
}

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
				if err := tsk.watching(); err != nil {
					fmt.Println("-----启动错误::", err)
					taskRunID := GenTaskID(tsk.ID)
					if Seched.Exists(taskRunID) {
						Seched.Remove(taskRunID)
					}
				}
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
		return fmt.Errorf("任务还未到开始时间")
	}

	if this.Endtime > 0 && this.Endtime < now {
		return fmt.Errorf("任务已结束")
	}
	return this.Run()
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

	if this.runnerBrowser == nil {
		this.runnerBrowser = make(map[int64]*bs.Browser)
	}

	if this.ID == 0 {
		return fmt.Errorf("任务未设置")
	}
	if this.Status == 0 {
		return fmt.Errorf("任务未启动")
	}
	if this.Callback == nil {
		this.Callback = func(str string) error {
			return nil
		}
		// return fmt.Errorf("未设置执行方法")
	}
	if this.OnClose == nil {
		this.OnClose = func() {
			fmt.Println("浏览器被关闭-----")
		}
	}

	taskID := GenTaskID(this.ID)
	if Seched.Exists(taskID) {
		return nil
	}
	if _, ok := dbTasks.Load(this.ID); !ok {
		dbTasks.Store(this.ID, this)
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
	fmt.Println("-------- 执行任务:", this.ID)
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
		taskDevices = append(taskDevices, &TaskClients{
			DeviceID:   0,
			DeviceType: 0,
		})
	}

	//设备需要分开启动,浏览器的和手机先隔离开
	var borwsers []int64
	var phones []int64

	for _, v := range taskDevices {
		switch v.DeviceType {
		case 0:
			borwsers = append(borwsers, v.DeviceID)
		case 1:
			phones = append(phones, v.DeviceID)
		}
	}

	wg := new(sync.WaitGroup)
	if len(borwsers) > 0 {
		fmt.Println("执行浏览器任务:", len(borwsers))
		wg.Go(func() {
			wgg := new(sync.WaitGroup)
			for _, bid := range borwsers { // 此处需要完善使用可控制输了的协程, sc_num
				wgg.Go(func() {
					ch := make(chan bool)
					this.runBrowser(bid, func(str string) error {
						if this.Callback != nil {
							return this.Callback(str)
						}
						return nil
					}, func() {
						if this.OnClose != nil {
							this.OnClose()
							ch <- true
						}
					}, func(str string) error {
						if this.OnUrlchange != nil {
							return this.OnUrlchange(str)
						}
						return nil
					})
					<-ch
				})
				time.Sleep(time.Duration(int64(funcs.RandomNumber(10, 200))))
			}
			wgg.Wait()
		})
	}
	if len(phones) > 0 {
		// this.runPhones(phones)
	}
	wg.Wait()

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

	if this.runnerBrowser != nil {
		for k, b := range this.runnerBrowser {
			b.Close()
			delete(this.runnerBrowser, k)
		}
	}
	return nil
}

func GenTaskID(id int64) string {
	return fmt.Sprintf("task-%d", id)
}

// 执行浏览器的任务
func (this *Task) runBrowser(
	browserID int64,
	callback func(string) error, //如果返回nil,则关闭浏览器,说明已经执行成功了
	closeback func(),
	urlchangeback func(string) error, // 如果不是nil,则关闭浏览器,说明url不允许执行
) (*bs.Browser, error) {
	brows, _ := bbs.New(browserID, bs.Options{
		Url:      this.DefUrl,
		JsStr:    this.ScriptStr,
		Headless: !(this.Headless == 1),
		Timeout:  time.Duration(time.Second * time.Duration(this.Timeout)),
		Ctx:      Seched.Ctx,
	})

	if closeback != nil {
		brows.OnClosed(func() {
			closeback()
		})
	}

	if callback != nil {
		brows.OnConsole(func(args []*runtime.RemoteObject) {
			for _, arg := range args {
				if arg.Value != nil {
					gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
					if gs.Get("type").String() == "kaka" {
						if err := callback(gs.Get("data").String()); err == nil {
							brows.Close()
						}
					}
				}
			}
		})
	}

	if urlchangeback != nil {
		brows.OnURLChange(func(url string) {
			if err := urlchangeback(url); err != nil {
				brows.Close()
			}
		})
	}

	brows.OpenBrowser()
	time.Sleep(time.Second * 1)

	if this.DefUrl != "" {
		brows.GoToUrl(this.DefUrl)
	}

	this.runnerBrowser[brows.ID] = brows

	go brows.RunJs(this.ScriptStr)
	return brows, nil
}

// 批量执行手机端任务
func (this *Task) runPhone(phoneIDs int64) {

}
