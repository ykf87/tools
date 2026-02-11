// task的人任务主要是记录落库，记录日志和发送ws
// 让系统可以显现的看见任务
// 任务分为 临时任务 和 常规任务。
// 任务还分为 周期任务 和 一次性任务。
// 如果遇到系统级的panic则再程序启动时将状态为正在执行的任务设置为执行失败
package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/listens/ws"
	"tools/runtimes/mainsignal"
	"tools/runtimes/scheduler"

	"gorm.io/gorm"
)

type Task struct {
	ID       int64                `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string               `json:"name" gorm:"index;not null"`     // 任务名称,用于查询任务：比如media_user处查询任务明细,一般用表名
	NameID   int64                `json:"name_id" gorm:"index;default:0"` // 对应的表ID,也是方便查询的
	Title    string               `json:"title" gorm:"not null;index"`    // 任务标题,对外显示的名称
	Addtime  int64                `json:"addtime" gorm:"index;not null"`  // 加入时间
	Endtime  int64                `json:"endtime" gorm:"index;default:0"` // 结束时间, 当设置了结束时间意味着任务结束了
	Limit    int                  `json:"limit" gorm:"default:0;index"`   // 并发限制,0 为不限制
	Nomal    int                  `json:"nomal" gorm:"index;default:0"`   // 是否临时任务,0为临时任务, 1为常规任务
	_sch     *scheduler.Scheduler `json:"-" gorm:"-"`                     // 调度器
	_mu      sync.Mutex           `json:"-" gorm:"-"`                     // 加锁,用于 _runners 的判断
	_runners map[string]*TaskRun  `json:"-" gorm:"-"`                     // 执行者列表
	db.BaseModel
}

// 任务执行者,只有在新建任务，结束任务落库
// 结束分为正常结束，错误结束，意外结束(panic)
type TaskRun struct {
	ID          int64                `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID      int64                `json:"task_id" gorm:"index;not null"`   // 任务表的ID
	RunID       string               `json:"run_id" gorm:"index;not null"`    // 执行的id,需要调用方设置,执行器通过这个id判断是否已有执行器
	Title       string               `json:"title" gorm:"index;not null"`     // 执行器标题
	Cycle       int64                `json:"cycle" gorm:"index;default:0"`    // 周期,单位秒
	AtTime      int64                `json:"at_time" gorm:"index;default:0"`  // 定点任务
	Status      int                  `json:"status" gorm:"index; default:0"`  // 任务状态，0-等待执行 1-正在执行, 2-执行完成. 在此处无法判断是否执行成功
	Total       float64              `json:"total" gorm:"default:0"`          // 任务总执行量
	Doned       float64              `json:"doned" gorm:"default:0"`          // 任务已执行量
	DoneMsg     string               `json:"done_msg" gorm:"default:null"`    // 执行完成消息
	StartAt     int64                `json:"start_at" gorm:"index;default:0"` // 开始时间,也就是任务开始执行的时间
	DoneTimes   int64                `json:"done_times" gorm:"default:0"`     // 执行次数
	EndAt       int64                `json:"end_at" gorm:"index;default:0"`   // 结束时间
	NextRuntime int64                `json:"next_runtime" gorm:"-"`           // 下一次执行时间
	_timeout    time.Duration        `json:"-" gorm:"-"`                      // 超时时间
	_sch        *scheduler.Scheduler `json:"-" gorm:"-"`                      // 调度器
	_runner     *scheduler.Runner    `json:"-" gorm:"-"`                      // 调度执行器
	db.BaseModel
}

// 任务执行的详细消息
type TaskRunMsg struct {
	TaskID      int64  `json:"task_id" gorm:"index;"`
	TaskRunID   int64  `json:"task_run_id" gorm:"index"`
	Msg         string `json:"msg" gorm:"not null"`           // 具体的消息
	Status      int    `json:"status" gorm:"default:0"`       // 状态, -1失败,0正在执行,1成功
	Addtime     int64  `json:"addtime" gorm:"not null"`       // 添加时间
	Tried       int    `json:"tried" gorm:"default:0"`        // 已重试次数
	NextRuntime int64  `json:"next_runtime" gorm:"default:0"` // 下一次重试时间
	Doned       int64  `json:"doned" gorm:"-"`                // 已执行次数
	db.BaseModel
}

var TaskActived sync.Map
var taskdb = db.TaskLogDB

func init() {
	taskdb.DB().AutoMigrate(&Task{})
	taskdb.DB().AutoMigrate(&TaskRun{})
	taskdb.DB().AutoMigrate(&TaskRunMsg{})

	// 系统启动的时候查询执行者的状态,如果是等待执行或者正在执行的强制置为执行失败
	taskdb.Write(func(tx *gorm.DB) error {
		return tx.Model(&TaskRun{}).Where("status = ? or status = ?", 0, 1).Updates(map[string]any{
			"status": -1,
			"end_at": time.Now().Unix(),
		}).Error
	})
	// taskdb.Model(&TaskRun{}).Where("status = ? or status = ?", 0, 1).Updates(map[string]any{
	// 	"status": -1,
	// 	"end_at": time.Now().Unix(),
	// })
	taskdb.Write(func(tx *gorm.DB) error {
		return tx.Model(&Task{}).Where("endtime = 0").Update("endtime", time.Now().Unix()).Error
	})
}

// 使用 name 和 nameId 作为唯一的组合
func getUniqueIndex(name string, id int64) string {
	if id == 0 {
		return name
	}
	return fmt.Sprintf("%s-%d", name, id)
}

// 发送所有任务到websocket
func SentAllTask() {
	TaskActived.Range(func(k, v any) bool {
		if tsk, ok := v.(*Task); ok {
			tsk.Sent()
		}
		return true
	})
}

// 创建一个任务,当tp或id为默认值时是临时任务
// name - 设置某一类的任务,建议使用调用方的表名,反正自己看,更多用于查询某个集合的任务
// id - 0 为临时的任务
// title - 任务名称
// limit - 该任务下执行器的执行并发数量
func NewTask(name string, id int64, title string, limit int, temp bool) (*Task, error) {
	tid := getUniqueIndex(name, id)
	if t, ok := TaskActived.Load(tid); ok {
		if tsk, ok := t.(*Task); ok {
			return tsk, nil
		}
	}
	tsk := new(Task)
	tsk.Name = name
	tsk.NameID = id
	tsk.Title = title
	tsk.Addtime = time.Now().Unix()
	tsk.Limit = limit
	if temp == true {
		tsk.Nomal = 0
	}
	if err := taskdb.Write(func(tx *gorm.DB) error {
		return tsk.Save(tsk, tx)
	}); err != nil {
		return nil, err
	}

	if limit > 0 {
		tsk._sch = scheduler.NewWithLimit(mainsignal.MainCtx, limit)
	} else {
		tsk._sch = scheduler.New(mainsignal.MainCtx)
	}

	tsk._runners = make(map[string]*TaskRun)
	TaskActived.Store(tid, tsk)

	if dt, err := config.Json.Marshal(map[string]any{
		"type": "task",
		"data": tsk,
	}); err == nil {
		ws.Broadcost(dt)
	}

	return tsk, nil
}

// 停止任务
func Stop(name string, id int64) {
	tid := getUniqueIndex(name, id)
	if t, ok := TaskActived.Load(tid); ok {
		if tsk, ok := t.(*Task); ok {
			for _, v := range tsk._runners {
				v.Stop()
			}
			tsk.Endtime = time.Now().Unix()
			taskdb.Write(func(tx *gorm.DB) error {
				return tsk.Save(tsk, tx)
			})
		}
	}
}

// 发送任务到ws
func (t *Task) Sent() {
	if dt, err := config.Json.Marshal(map[string]any{
		"type": "task",
		"data": t,
	}); err == nil {
		ws.Broadcost(dt)
		for _, v := range t._runners {
			v.fullMsg()

			lastsmg := new(TaskRunMsg)
			err := taskdb.DB().Model(&TaskRunMsg{}).Where("task_id = ? and task_run_id = ?", t.ID, v.ID).Order("addtime DESC").First(lastsmg).Error
			if err == nil {
				lastsmg.sent(v)
			} else {
				trm := &TaskRunMsg{
					TaskID:      v.TaskID,
					TaskRunID:   v.ID,
					Status:      0,
					Addtime:     time.Now().Unix(),
					Msg:         "任务等待中...",
					Tried:       0,
					NextRuntime: v._runner.GetNextRunTime().Unix(),
				}
				trm.sent(v)
			}
		}
	}
}

// 添加执行的子任务
// 子任务必须设置id,用于重复添加的限制
func (t *Task) AddChild(runnerID, title string, timeout time.Duration) (*TaskRun, error) {
	t._mu.Lock()
	defer t._mu.Unlock()
	if tr, ok := t._runners[runnerID]; ok {
		return tr, nil
	}

	tr := new(TaskRun)
	tr.TaskID = t.ID
	tr.RunID = runnerID
	tr.Title = title
	tr.StartAt = time.Now().Unix()
	tr._timeout = timeout
	tr._sch = t._sch

	if err := taskdb.Write(func(tx *gorm.DB) error {
		return tr.Save(tr, tx)
	}); err != nil {
		return nil, err
	}
	t._runners[runnerID] = tr
	tr.fullMsg()

	return tr, nil
}

// 发送完整的TaskRunner消息
func (tr *TaskRun) fullMsg() {
	if tr._runner != nil {
		tr.NextRuntime = tr._runner.GetNextRunTime().Unix()
	}
	if dt, err := config.Json.Marshal(map[string]any{
		"type": "task-runner",
		"data": tr,
	}); err == nil {
		ws.Broadcost(dt)
	}
}

// 发送执行消息
func (trm *TaskRunMsg) sent(tr *TaskRun) {
	trm.NextRuntime = tr._runner.GetNextRunTime().Unix()
	trm.Doned = int64(tr._runner.GetRunTimes())
	if dt, err := config.Json.Marshal(map[string]any{
		"type": "task-runner-msg",
		"data": trm,
	}); err == nil {
		ws.Broadcost(dt)
	}
}

// 停止任务
func (tr *TaskRun) Stop() {
	tr._runner.Stop()
}

// 启动定时任务, interval - 单位是秒
func (tr *TaskRun) StartInterval(interval int64, callback func(*TaskRun) error) error {
	if callback == nil {
		return errors.New(i18n.T("Please set callabck func"))
	}
	tr.Cycle = interval
	taskdb.Write(func(tx *gorm.DB) error {
		return tr.Save(tr, tx)
	})

	tr.setRunner(callback)._runner.Every(time.Second * time.Duration(interval))
	tr.runnerCallback()
	tr._runner.RunNow()
	return nil
}

// 启动定点任务, 比如19:20:48启动
func (tr *TaskRun) StartAtTime(timer int64, callback func(*TaskRun) error) error {
	if callback == nil {
		return errors.New(i18n.T("Please set callabck func"))
	}
	h, m, s := funcs.MsToHMS(timer)
	tr.AtTime = timer
	taskdb.Write(func(tx *gorm.DB) error {
		return tr.Save(tr, tx)
	})

	tr.setRunner(callback)._runner.DailyRandomAt(h, m, s, 5, nil)
	tr.runnerCallback()
	tr._runner.Run()
	return nil
}

// 设置执行器
func (tr *TaskRun) setRunner(callback func(*TaskRun) error) *TaskRun {
	tr._runner = tr._sch.NewRunner(func(ctx context.Context) error {
		if tr.StartAt < 1 {
			tr.StartAt = time.Now().Unix()
			tr.Status = 1
			taskdb.Write(func(tx *gorm.DB) error {
				return tr.Save(tr, tx)
			})
		}
		trm := &TaskRunMsg{
			TaskID:      tr.TaskID,
			TaskRunID:   tr.ID,
			Status:      1,
			Addtime:     time.Now().Unix(),
			Msg:         "开始执行...",
			Tried:       tr._runner.GetTryTimers(),
			NextRuntime: tr._runner.GetNextRunTime().Unix(),
			Doned:       int64(tr._runner.GetRunTimes()),
		}
		trm.sent(tr)
		return callback(tr)
	}, tr._timeout, mainsignal.MainCtx)

	trm := &TaskRunMsg{
		TaskID:      tr.TaskID,
		TaskRunID:   tr.ID,
		Status:      0,
		Addtime:     time.Now().Unix(),
		Msg:         "任务等待中...",
		Tried:       0,
		NextRuntime: tr._runner.GetNextRunTime().Unix(),
	}
	trm.sent(tr)
	return tr
}

// 设置执行器回调
func (tr *TaskRun) runnerCallback() {
	if tr._runner != nil {
		tr._runner.SetCloser(func() {
			tr.EndAt = time.Now().Unix()
			tr.Status = 2
			taskdb.Write(func(tx *gorm.DB) error {
				return tr.Save(tr, tx)
			})
			tr.fullMsg()

			taskdb.Write(func(tx *gorm.DB) error {
				return tx.Model(&Task{}).Where("id = ?", tr.TaskID).Update("endtime", tr.EndAt).Error
			})
		}).SetOnceDone(func(tried int32, err error, nextRunTime time.Time) {
			var status int
			var msg string
			if err == nil {
				msg = i18n.T("Task run success")
				status = 1
			} else {
				status = -1
				msg = i18n.T("Task run error: %s", err.Error())
			}
			tr.DoneTimes = int64(tr._runner.GetRunTimes())
			taskdb.Write(func(tx *gorm.DB) error {
				return tr.Save(tr, tx)
			})
			// fmt.Println(rtt, "------ 报错执行次数出错")
			// go tr.Save(tr, taskdb)
			trm := &TaskRunMsg{
				TaskID:      tr.TaskID,
				TaskRunID:   tr.ID,
				Status:      status,
				Addtime:     time.Now().Unix(),
				Msg:         msg,
				Tried:       int(tried),
				NextRuntime: nextRunTime.Unix(),
			}
			// go tr.fullMsg()
			taskdb.Write(func(tx *gorm.DB) error {
				return trm.Save(trm, tx)
			})
			trm.sent(tr)
		}).SetError(func(err error, tried int32) {
			trm := &TaskRunMsg{
				TaskID:      tr.TaskID,
				TaskRunID:   tr.ID,
				Status:      -1,
				Addtime:     time.Now().Unix(),
				Msg:         fmt.Sprintf("%s, 重试:%d", err.Error(), tried),
				Tried:       int(tried),
				NextRuntime: tr._runner.GetNextRunTime().Unix(),
			}
			taskdb.Write(func(tx *gorm.DB) error {
				return trm.Save(trm, tx)
			})
			trm.sent(tr)
		}).SetMaxTry(5).SetRetryDelay(time.Second * 60)
	}
}

// 获取重试次数
func (tr *TaskRun) GetTried() int {
	return tr._runner.GetTryTimers()
}

// 上报任务进度
func (tr *TaskRun) ReportSchedule(total, doned float64) {
	tr.Total = total
	tr.Doned = doned
	tr.fullMsg()
}

// 设置子任务停止时间
func (tr *TaskRun) StopAt(t time.Time) {
	tr._runner.SetStopAt(t)
}

// 获取执行器上下文
func (tr *TaskRun) GetCtx() context.Context {
	return tr._runner.GetCtx()
}
