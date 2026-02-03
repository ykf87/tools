package tasklog

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/clients"
	"tools/runtimes/mainsignal"
	"tools/runtimes/scheduler"
)

// Task 并不是写入数据库的,而是作为任务载体发送到websocket中
// 用于任务日志管理和ws
type Task struct {
	TaskID  string                 `json:"task_id"`  // 前端使用此字段作为判断,TaskID在每次发起任务时通过调用方设定,也是作为是否唯一的判断
	Title   string                 `json:"title"`    // 任务名称
	BeginAt int64                  `json:"begin_at"` // 开始于什么时间
	MaxCh   int                    `json:"max_ch"`   // 并发控制
	sch     *scheduler.Scheduler   `json:"-"`        // 调度器
	mu      sync.RWMutex           `json:"-"`        // 读写锁
	runners map[string]*TaskRunner `json:"-"`        // 执行的任务
}

// 具体执行的任务
type TaskRunner struct {
	ID       string               `json:"id"`       // 调用方设置的id
	Title    string               `json:"title"`    // 执行的标题
	StartAt  int64                `json:"start_at"` // 开始执行时间
	Runner   *scheduler.Runner    `json:"-"`        // 执行器
	CallBack func(string) error   `json:"-"`        // 内容回调
	opt      any                  `json:"-"`        //配置
	msg      chan string          `json:"-"`        // 消息接收器
	sch      *scheduler.Scheduler `json:"-"`        // 调度器,和Task调度器一致
	Bss      *bs.Browser          `json:"-"`        // 浏览器
	Phone    clients.Phone        `json:"-"`        // 手机端
	ctx      context.Context      `json:"-"`        //上下文
	// callback scheduler.TaskFunc   `json:"-"`        // 执行回调
	// onerror  scheduler.ErrFun     `json:"-"`        // 错误回调
	// onclose  scheduler.CloseFun   `json:"-"`        // 结束回调
}

type TaskLog struct {
	ID      int64  `json:"id" gorm:"primaryKey;autoIncrement"`  // 表自增id
	TaskID  int64  `json:"task_id" gorm:"index;default:0"`      // 任务对于的表的id号,0代表任务并不是从数据库发起的
	RunerID string `json:"runer_id" gorm:"not null"`            // 执行的id
	Title   string `json:"title" gorm:"type:varchar(32);index"` // 任务名称
	Tag     string `json:"tag" gorm:"index;type:varchar(32);"`  // 任务标签
}

var RunnerTask sync.Map

// 创建任务
// taskID 唯一的任务id
// title 任务名称
// maxch 并发控制
// jitter 并发随机时间,秒
func NewTaskLog(taskID, title string, maxch, jitter int) *Task {
	if to, ok := RunnerTask.Load(taskID); ok {
		if t, ok := to.(*Task); ok {
			return t
		}
	}
	t := &Task{
		TaskID:  taskID,
		Title:   title,
		BeginAt: time.Now().Unix(),
		MaxCh:   maxch,
		runners: make(map[string]*TaskRunner),
	}
	if maxch > 0 {
		t.sch = scheduler.NewWithLimit(mainsignal.MainCtx, maxch)
		t.sch.SetJitter(time.Second * time.Duration(jitter))
	} else {
		t.sch = scheduler.New(mainsignal.MainCtx)
	}

	return t
}

// 添加周期执行任务
func (t *Task) Append(ctx context.Context, runid, title string, callback func(string) error) *TaskRunner {
	t.mu.Lock()
	defer t.mu.Unlock()
	if rr, ok := t.runners[runid]; ok {
		return rr
	}

	tr := &TaskRunner{
		sch:      t.sch,
		ID:       runid,
		Title:    title,
		ctx:      ctx,
		msg:      make(chan string),
		CallBack: callback,
	}
	t.runners[runid] = tr
	return tr
}

// 添加回调函数
// func (tr *TaskRunner) SetCallback(callback scheduler.TaskFunc) *TaskRunner {
// 	tr.callback = callback
// 	return tr
// }
// func (tr *TaskRunner) SetError(fun scheduler.ErrFun) *TaskRunner {
// 	tr.onerror = fun
// 	return tr
// }
// func (tr *TaskRunner) SetClose(fun scheduler.CloseFun) *TaskRunner {
// 	tr.onclose = fun
// 	return tr
// }

func (tr *TaskRunner) callweb(ctx context.Context) error {
	opt, ok := tr.opt.(*bs.Options)
	if !ok {
		return fmt.Errorf("option error")
	}
	opt.Ctx = ctx
	bss, err := bs.BsManager.New(opt.ID, opt, true)
	if err != nil {
		return err
	}

	bss.Opts.Msg = tr.msg
	tr.Bss = bss

	if err := tr.Bss.OpenBrowser(); err != nil {
		return err
	}

	if opt.Url != "" {
		bss.GoToUrl(opt.Url)
	}
	go bss.RunJs(opt.JsStr)
	// fmt.Println(opt.JsStr, "====")

	myctx, mycancle := context.WithCancel(ctx)
	defer mycancle()

	for {
		select {
		case msg := <-tr.msg:
			if err := tr.CallBack(msg); err != nil {
				return err
			}
		case <-myctx.Done():
			goto CLOSER
		}
	}
CLOSER:
	return nil
}

// 启动执行任务
// ctx上下文
// browserID 浏览器执行id
func (t *TaskRunner) StartWeb(opt *bs.Options, timeout time.Duration) error {
	if opt.Ctx == nil {
		opt.Ctx = t.ctx
	}
	t.opt = opt

	t.Runner = t.sch.NewRunner(t.callweb, timeout, t.ctx)

	return nil
}

func (t *TaskRunner) StartPhone(ctx context.Context, timeout time.Duration) {

}
