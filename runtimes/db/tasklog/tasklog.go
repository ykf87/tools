package tasklog

import (
	"context"
	"fmt"
	"sync"
	"time"
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
	IsTemp  bool                   `json:"is_temp"`  // 是否是临时任务
	sch     *scheduler.Scheduler   `json:"-"`        // 调度器
	mu      sync.RWMutex           `json:"-"`        // 读写锁
	runners map[string]*TaskRunner `json:"-"`        // 执行的任务
	ctx     context.Context        `json:"-"`        // 上下文
	cancle  context.CancelFunc     `json:"-"`        // 关闭
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
func NewTaskLog(taskID, title string, maxch, jitter int, temp bool) *Task {
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
		IsTemp:  temp,
	}

	t.ctx, t.cancle = context.WithCancel(mainsignal.MainCtx)

	if maxch > 0 {
		t.sch = scheduler.NewWithLimit(t.ctx, maxch)
		t.sch.SetJitter(time.Second * time.Duration(jitter))
	} else {
		t.sch = scheduler.New(t.ctx)
	}

	RunnerTask.Store(taskID, t)
	t.Sent()
	return t
}

// 添加周期执行任务
// 回调函数接收 *TaskRunner, 设置内容并向 msg chan string 发送消息才算完成一次回调
func (t *Task) Append(ctx context.Context, runid, title string, callback func(string, *TaskRunner) error) *TaskRunner {
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
		Msg:      make(chan string),
		callBack: callback,
		taskID:   t.TaskID,
	}

	tr.ctx, tr.cancle = context.WithCancel(t.ctx)

	go func() {
		for {
			select {
			case msg := <-tr.Msg:
				go tr.Sent(msg)
			case msg := <-tr.ErrMsg:
				go tr.SentErr(msg)
			case <-tr.ctx.Done():
				tr.endAt = time.Now().Unix()
				go tr.Sent("任务结束")
				return
			}
		}
	}()

	t.runners[runid] = tr
	tr.Sent("任务等待开始...")
	return tr
}

// 获取所有在执行的任务
func GetRuningTasks() []*TaskWs {
	var tsk []*TaskWs
	RunnerTask.Range(func(k, v any) bool {
		fmt.Println("有任务在哦,", k)
		if t, ok := v.(*Task); ok {
			if tskk := t.Sent(); tskk != nil {
				tsk = append(tsk, tskk)
				t.mu.Lock()
				for _, v := range t.runners {
					if v.Runner.IsRuning() {
						v.Sent("")
					}
				}
				t.mu.Unlock()
			}
		}
		return true
	})
	return tsk
}
