// golang后台执行的任务
package tasks

import (
	"context"
	"fmt"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/listens/ws"

	"gorm.io/gorm"
)

type DeviceType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var Types = []DeviceType{
	{
		ID:   0,
		Name: "Web端",
	},
	{
		ID:   1,
		Name: "手机端",
	},
}

// 任务表
type Task struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Title     string `json:"title" gorm:"index;not null;type:varchar(32)" form:"title"`          // 任务名称
	Tp        int    `json:"type" gorm:"index;default:0" form:"type"`                            // 任务类型,分2种, 0-web端  1-手机端
	Starttime int64  `json:"starttime" gorm:"index;default:0" form:"starttime" parse:"datetime"` // 任务开始时间
	Endtime   int64  `json:"endtime" gorm:"index;default:0" form:"endtime" parse:"datetime"`     // 任务结束时间
	Status    int    `json:"status" gorm:"type:tinyint(1);default:1;index" form:"status"`        // 任务状态, 1-可执行 0-不可执行
	Errmsg    string `json:"errmsg" gorm:"default:null" form:"errmsg"`                           // 错误信息
	AdminId   int64  `json:"admin_id" gorm:"index;not null"`                                     // 管理员id
	Cycle     int64  `json:"cycle" gorm:"default:0" form:"cycle"`                                // 任务周期,单位秒,0为不重复执行,大于0表示间隔多久自动重复执行
	RetryMax  int    `json:"retry_max" gorm:"default:0"`                                         // 最大重试次数
	Timeout   int64  `json:"timeout" gorm:"default:0"`                                           // 单次超时（秒）
	Priority  int    `json:"priority" gorm:"default:0"`                                          // 优先级
	CatchUp   bool   `json:"catch_up" gorm:"default:false"`                                      // 补跑漏掉的周期
}

// 任务执行表
type TaskRun struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    int64  `json:"task_id" gorm:"index;not null"`
	RunStatus int    `json:"run_status" gorm:"index;default:0"` // 0-表示执行中 1-执行成功 -1-执行失败
	StatusMsg string `json:"status_msg" gorm:"default:null"`    // 执行结果
	RunTime   int64  `json:"run_time" gorm:"index;not null"`    // 本次执行开始时间
	DoneTime  int64  `json:"done_time" gorm:"index;default:0"`  // 本次执行完成时间
}

// TaskLog 任务执行详细日志表
type TaskLog struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    int64  `json:"task_id" gorm:"index;not null"`
	TaskRunID int64  `json:"task_run_id" gorm:"index;not null"`
	Addtime   int64  `json:"addtime" gorm:"index;not null"`
	LogStatus int    `json:"log_status" gorm:"index;default:0"`
	Msg       string `json:"msg" gorm:"default:null"`
}

// 任务设备表
type TaskClients struct {
	TaskID     int64 `json:"task_id" gorm:"primaryKey"`     // 任务ID
	DeviceID   int64 `json:"device_id" gorm:"primaryKey"`   // 设备id,在clients下的 browsers 或者 phones 表
	DeviceType int   `json:"device_type" gorm:"primaryKey"` // 设备类型, 0-web端 1-手机端
}

// 任务标签表
type TaskTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"index;not null"`
}

// 任务对于的标签表
type TaskToTag struct {
	TaskID int64 `json:"task_id" gorm:"primaryKey"`
	TagID  int64 `json:"tag_id" gorm:"primaryKey"`
}

// 临时任务
type TempTask struct {
	ID        int64
	Title     string
	Timeout   int64
	Priority  int
	CreatedAt int64
}

func init() {
	db.TaskDB.AutoMigrate(&Task{})
	db.TaskDB.AutoMigrate(&TaskClients{})
	db.TaskDB.AutoMigrate(&TaskTag{})
	db.TaskDB.AutoMigrate(&TaskToTag{})
	db.TaskDB.AutoMigrate(&TaskRun{})
	db.TaskDB.AutoMigrate(&TaskLog{})

	// 启动任务监听
	InitScheduler(
		func() ([]Task, error) {
			var tasks []Task
			err := dbs.Where("status = 1").Find(&tasks).Error
			return tasks, err
		},
		func(ctx context.Context, task *Task, runID int64) error {
			// log.Println("[run task] 执行任务:", task.ID)
			for {
				select {
				case <-ctx.Done():
					// ⚠️ 这是“正常的被中断结束”
					return ctx.Err()
				default:
					// log.Println("----执行代码")
					return nil
					// doOneStep()
				}
			}
			// return ctx.Err()
			// return nil
		},
	)
}

func (this *Task) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.TaskDB
	}
	defer NotifyTaskChanged(this.ID)
	if this.ID > 0 {
		return tx.Model(&Task{}).Where("id = ?", this.ID).
			Updates(map[string]any{
				"title":     this.Title,
				"tp":        this.Tp,
				"starttime": this.Starttime,
				"endtime":   this.Endtime,
				"status":    this.Status,
				"errmsg":    this.Errmsg,
				"admin_id":  this.AdminId,
				"cycle":     this.Cycle,
				"retry_max": this.RetryMax,
				"timeout":   this.Timeout,
				"priority":  this.Priority,
				"catch_up":  this.CatchUp,
			}).Error
	} else {
		if this.Starttime < 1 {
			this.Starttime = time.Now().Unix()
		}
		err := tx.Create(this).Error
		if err == nil {
			ws.SentBus(this.AdminId, "task", this, "")
		}
		return err
	}
}

// 获取任务总数
func GetTotalTask(groupname string, adminid int64) int64 {
	var total int64
	md := db.TaskDB.Model(&Task{}).Where("admin_id = ?", adminid)
	if groupname != "" {
		md.Where("group_name = ?", groupname)
	}
	md.Count(&total)
	return total
}

// 获取分组的任务
func GetTasks(page, limit int, query string, adminid int64) ([]*Task, int64) {
	var tks []*Task
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	md := db.TaskDB.Model(&Task{}).Where("admin_id = ?", adminid)
	if query != "" {
		qs := fmt.Sprintf("%%%s%%", query)
		md.Where("title like ?", qs)
	}

	var total int64
	md.Count(&total)

	md.Order("starttime DESC").Offset((page - 1) * limit).Limit(limit).Find(&tks)
	return tks, total
}

// 获取tags
func GetTags() []*TaskTag {
	var tgs []*TaskTag
	db.TaskDB.Model(&TaskTag{}).Find(&tgs)
	return tgs
}
