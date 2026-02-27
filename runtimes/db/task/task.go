// task的人任务主要是记录落库，记录日志和发送ws
// 让系统可以显现的看见任务
// 任务分为 临时任务 和 常规任务。
// 任务还分为 周期任务 和 一次性任务。
// 如果遇到系统级的panic则再程序启动时将状态为正在执行的任务设置为执行失败
package task

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"
	"tools/runtimes/listens/ws"
	"tools/runtimes/mainsignal"
	"tools/runtimes/sch"

	"gorm.io/gorm"
)

type Task struct {
	ID       int64               `json:"id" gorm:"primaryKey;autoIncrement"`
	Tid      string              `json:"tid" gorm:"index"`
	Name     string              `json:"name" gorm:"index;not null"`     // 任务名称,用于查询任务：比如media_user处查询任务明细,一般用表名
	NameID   int64               `json:"name_id" gorm:"index;default:0"` // 对应的表ID,也是方便查询的
	Title    string              `json:"title" gorm:"not null;index"`    // 任务标题,对外显示的名称
	Addtime  int64               `json:"addtime" gorm:"index;not null"`  // 加入时间
	Endtime  int64               `json:"endtime" gorm:"index;default:0"` // 结束时间, 当设置了结束时间意味着任务结束了
	Limit    int                 `json:"limit" gorm:"default:0;index"`   // 并发限制,0 为不限制
	Nomal    int                 `json:"nomal" gorm:"index;default:0"`   // 是否临时任务,0为临时任务, 1为常规任务
	_sch     *sch.Scheduler      `json:"-" gorm:"-"`                     // 调度器
	_mu      sync.Mutex          `json:"-" gorm:"-"`                     // 加锁,用于 _runners 的判断
	_runners map[string]*TaskRun `json:"-" gorm:"-"`                     // 执行者列表
	db.BaseModel
}

// 任务执行者,只有在新建任务，结束任务落库
// 结束分为正常结束，错误结束，意外结束(panic)
type TaskRun struct {
	ID          int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID      int64          `json:"task_id" gorm:"index;not null"`   // 任务表的ID
	TaskTid     string         `json:"task_tid" gorm:"index"`           // 用于查找已存在的任务
	TaskName    string         `json:"task_name"`                       // 总任务名称
	RunID       string         `json:"run_id" gorm:"index;not null"`    // 执行的id,需要调用方设置,执行器通过这个id判断是否已有执行器
	Title       string         `json:"title" gorm:"index;not null"`     // 执行器标题
	Cycle       int64          `json:"cycle" gorm:"index;default:0"`    // 周期,单位秒
	AtTime      int64          `json:"at_time" gorm:"index;default:0"`  // 定点任务
	Status      int            `json:"status" gorm:"index; default:0"`  // 任务状态，0-等待执行 1-正在执行, 2-执行完成. 在此处无法判断是否执行成功
	Total       float64        `json:"total" gorm:"default:0"`          // 任务总执行量
	Doned       float64        `json:"doned" gorm:"default:0"`          // 任务已执行量
	DoneMsg     string         `json:"done_msg" gorm:"default:null"`    // 执行完成消息
	StartAt     int64          `json:"start_at" gorm:"index;default:0"` // 开始时间,也就是任务开始执行的时间
	DoneTimes   int64          `json:"done_times" gorm:"default:0"`     // 执行次数
	EndAt       int64          `json:"end_at" gorm:"index;default:0"`   // 结束时间
	NextRuntime int64          `json:"next_runtime" gorm:"-"`           // 下一次执行时间
	_timeout    time.Duration  `json:"-" gorm:"-"`                      // 超时时间
	_sch        *sch.Scheduler `json:"-" gorm:"-"`                      // 调度器
	_runner     *sch.Runner    `json:"-" gorm:"-"`                      // 调度执行器
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
	TaskName    string `json:"task_name"`                     // 大任务名称
	RunID       string `json:"run_id"`                        // 子任务id
	TaskTid     string `json:"task_tid" gorm:"index"`         // 大任务唯一标识
	db.BaseModel
}

var TaskActived sync.Map
var taskdb = db.TaskLogDB

func init() {
	taskdb.DB().AutoMigrate(&Task{})
	taskdb.DB().AutoMigrate(&TaskRun{})
	taskdb.DB().AutoMigrate(&TaskRunMsg{})

	// 系统启动的时候查询执行者的状态,如果是等待执行或者正在执行的强制置为执行失败
	go func() {
		deleteTimeoutTaskRunAndMsg(time.Now().UTC().AddDate(0, 0, -7))

		for {
			select {
			case <-time.After(24 * time.Hour):
				deleteTimeoutTaskRunAndMsg(time.Now().UTC().AddDate(0, 0, -7))
			case <-mainsignal.MainCtx.Done():
				return
			}
		}
	}()
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

// 定时删除任务消息
func deleteTimeoutTaskRunAndMsg(timer time.Time) {
	taskdb.Write(func(tx *gorm.DB) error {
		now := timer.Unix()
		var taskIds []int64
		if err := tx.Model(&Task{}).Select("id").Where("endtime > ? and addtime <= ?", 0, now).Find(&taskIds).Error; err != nil {
			return err
		}
		if err := tx.Where("endtime > ? and addtime <= ?", 0, now).Delete(&Task{}).Error; err != nil {
			return err
		}

		taskRunModel := tx.Where("start_at <= ?", now)
		if len(taskIds) > 0 {
			taskRunModel = taskRunModel.Or("task_id in ?", taskIds)
		}
		if err := taskRunModel.Delete(&TaskRun{}).Error; err != nil {
			return err
		}

		taskRunMsgModel := tx.Where("start_at <= ?", now)
		if len(taskIds) > 0 {
			taskRunMsgModel = taskRunMsgModel.Or("task_id in ?", taskIds)
		}
		if err := taskRunMsgModel.Delete(&TaskRunMsg{}).Error; err != nil {
			return err
		}
		return nil
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
	tsk.Tid = tid
	if temp == true {
		tsk.Nomal = 0
	}
	if err := taskdb.Write(func(tx *gorm.DB) error {
		return tsk.Save(tsk, tx)
	}); err != nil {
		return nil, err
	}

	tsk._sch = sch.NewScheduler(limit)

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
				v.SentMsg("任务等待中...", 0, false)
			}
		}
	}
}

// 添加定时子任务
func (t *Task) AddInterval(
	id string, // 唯一的id
	title string, // 标题
	interval time.Duration, // 执行间隔
	timeout time.Duration, // 超时时间
	retry int, // 重试次数
	retryDelay time.Duration, // 重试间隔
	expireAt time.Time, // ⭐ 新增,在这个时间点停止任务
	job func(*TaskRun) error, // 执行的方法
) (*TaskRun, error) {
	t._mu.Lock()
	defer t._mu.Unlock()
	if tr, ok := t._runners[id]; ok {
		return tr, nil
	}

	tr := new(TaskRun)
	tr.TaskID = t.ID
	tr.RunID = id
	tr.Title = title
	tr.StartAt = time.Now().Unix()
	tr._timeout = timeout
	tr._sch = t._sch
	tr.TaskName = t.Name
	tr.TaskTid = t.Tid

	if err := taskdb.Write(func(tx *gorm.DB) error {
		return tr.Save(tr, tx)
	}); err != nil {
		return nil, err
	}

	r, err := t._sch.AddInterval(
		id,
		interval,
		timeout,
		retry,
		retryDelay,
		expireAt,
		0.18,
		func(context.Context) error {
			tr.SentMsg("开始执行...", 1, true)
			return job(tr)
		},
		tr.onComplate,
		tr.onClose,
	)
	if err != nil {
		return nil, err
	}

	tr._runner = r
	t._runners[id] = tr
	tr.fullMsg()
	tr.SentMsg("任务等待中...", 0, false)

	return tr, nil
}

func (t *Task) AddCron(
	id string, // 唯一的id
	title string, // 标题
	timer int64, // 可以转成时分秒的时间差，如:-28800000
	timeout time.Duration, // 超时时间
	retry int, // 重试次数
	retryDelay time.Duration, // 重试间隔
	expireAt time.Time, // ⭐ 新增,在这个时间点停止任务
	job func(*TaskRun) error, // 执行的方法
) (*TaskRun, error) {
	t._mu.Lock()
	defer t._mu.Unlock()
	if tr, ok := t._runners[id]; ok {
		return tr, nil
	}

	tr := new(TaskRun)
	tr.TaskID = t.ID
	tr.RunID = id
	tr.Title = title
	tr.StartAt = time.Now().Unix()
	tr._timeout = timeout
	tr._sch = t._sch
	tr.TaskName = t.Name
	tr.TaskTid = t.Tid

	if err := taskdb.Write(func(tx *gorm.DB) error {
		return tr.Save(tr, tx)
	}); err != nil {
		return nil, err
	}

	h, m, s := funcs.MsToHMS(timer)
	// fmt.Println("启动时间: ", h, m, s, "---", timer)
	r, err := t._sch.AddCron(
		id,
		fmt.Sprintf("%d %d %d * * *", s, m, h),
		timeout,
		retry,
		retryDelay,
		expireAt,
		0.043,
		func(context.Context) error {
			tr.SentMsg("开始执行...", 1, true)
			return job(tr)
		},
		tr.onComplate,
		tr.onClose,
	)
	if err != nil {
		return nil, err
	}

	tr._runner = r
	t._runners[id] = tr
	tr.fullMsg()
	tr.SentMsg("任务等待中...", 0, false)

	return tr, nil
}

func (tr *TaskRun) onComplate(id string, err error) {
	var status int
	var msg string
	if err == nil {
		msg = i18n.T("本次任务执行完成")
		status = 1
	} else {
		status = -1
		msg = i18n.T("Task run error: %s", err.Error())
	}
	tr.DoneTimes = int64(tr._runner.RunCount())
	taskdb.Write(func(tx *gorm.DB) error {
		return tr.Save(tr, tx)
	})
	tr.SentMsg(msg, status, true)
}

func (tr *TaskRun) onClose(id string) {
	tr.EndAt = time.Now().Unix()
	tr.Status = 2
	taskdb.Write(func(tx *gorm.DB) error {
		return tr.Save(tr, tx)
	})
	tr.fullMsg()

	taskdb.Write(func(tx *gorm.DB) error {
		return tx.Model(&Task{}).Where("id = ?", tr.TaskID).Update("endtime", tr.EndAt).Error
	})
}

// 发送完整的TaskRunner消息
func (tr *TaskRun) fullMsg() {
	if toc, ok := TaskActived.Load(tr.TaskTid); ok {
		if t, ok := toc.(*Task); ok {
			if _, ok := t._runners[tr.RunID]; ok {
				if tr._runner != nil {
					tr.NextRuntime = tr._runner.NextRunTime().Unix()
				}
				if dt, err := config.Json.Marshal(map[string]any{
					"type": "task-runner",
					"data": tr,
				}); err == nil {
					ws.Broadcost(dt)
				}
			}
		}
	}
}

// 发送结束信号
func (tr *TaskRun) RemoveMsg() {
	if dt, err := config.Json.Marshal(map[string]any{
		"type": "task-runner-remove",
		"data": tr,
	}); err == nil {
		ws.Broadcost(dt)
	}
}

// 发送执行消息
func (trm *TaskRunMsg) sent(tr *TaskRun) {
	if toc, ok := TaskActived.Load(tr.TaskTid); ok {
		if t, ok := toc.(*Task); ok {
			if _, ok := t._runners[tr.RunID]; ok {
				trm.NextRuntime = tr._runner.NextRunTime().Unix()
				trm.Doned = tr._runner.RunCount()
				trm.TaskName = tr.TaskName
				trm.RunID = tr.RunID
				if dt, err := config.Json.Marshal(map[string]any{
					"type": "task-runner-msg",
					"data": trm,
				}); err == nil {
					ws.Broadcost(dt)
				}
			}
		}
	}
}

// 停止任务
func (tr *TaskRun) Stop() {
	tr._sch.Remove(tr._runner.GetID())
	// tr._runner.Stop()
}

// 获取重试次数
func (tr *TaskRun) GetTried() int {
	return int(tr._runner.LastRetryCount())
}

// 上报任务进度
func (tr *TaskRun) ReportSchedule(total, doned float64) {
	tr.Total = total
	tr.Doned = doned
	tr.fullMsg()
}

// 设置子任务停止时间
func (tr *TaskRun) StopAt(t time.Time) {
	tr._runner.SetExpireAt(t)
	// tr._runner.SetStopAt(t)
}

// 设置开始时间
func (tr *TaskRun) SetStartAt(t time.Time) {
	tr._runner.SetStartAt(t)
}

// 获取执行器上下文
func (tr *TaskRun) GetCtx() context.Context {
	return tr._runner.GetCtx()
}

// 马上执行
func (tr *TaskRun) RunNow() {
	tr._sch.RunNow(tr.RunID)
}

// 发送消息给前端
func (tr *TaskRun) SentMsg(msg string, status int, todb bool) {
	trm := &TaskRunMsg{
		TaskID:      tr.TaskID,
		TaskRunID:   tr.ID,
		TaskTid:     tr.TaskTid,
		Status:      status,
		Addtime:     time.Now().Unix(),
		Msg:         msg,
		TaskName:    tr.TaskName,
		RunID:       tr.RunID,
		Tried:       int(tr._runner.LastRetryCount()),
		NextRuntime: tr._runner.NextRunTime().Unix(),
	}
	trm.sent(tr)
	if todb == true {
		taskdb.Write(func(tx *gorm.DB) error {
			return trm.Save(trm, tx)
		})
	}
}

// 查询任务
func (t *Task) Query(runid string) *TaskRun {
	t._mu.Lock()
	defer t._mu.Unlock()
	if t, ok := t._runners[runid]; ok {
		return t
	}
	return nil
}

// 停止任务
func (t *Task) Stop(runid string) {
	if tr := t.Query(runid); tr != nil {
		t._sch.Remove(runid)
		t._mu.Lock()
		delete(t._runners, runid)
		t._mu.Unlock()
		tr.RemoveMsg()
	}
}

func (t *Task) RemoveMsg() {
	if dt, err := config.Json.Marshal(map[string]any{
		"type": "task-remove",
		"data": t.Tid,
	}); err == nil {
		ws.Broadcost(dt)
	}
}

// 停止所有任务
func (t *Task) StopAll() {
	t._sch.Stop()
	t._mu.Lock()
	t._runners = make(map[string]*TaskRun)
	t._mu.Unlock()
	t.RemoveMsg()
	TaskActived.Delete(t.Tid)
}
