// golang后台执行的任务
package tasks

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/db/jses"
	"tools/runtimes/listens/ws"
	"tools/runtimes/scheduler"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var RunnerTasks sync.Map

type DeviceType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var Types = []DeviceType{
	{
		ID:   0,
		Name: "浏览器",
	},
	{
		ID:   1,
		Name: "手机端",
	},
}

// type deviceList struct {
// 	ID    int64  `json:"id"`
// 	Name  string `json:"name"`
// 	Num   int64  `json:"num"`   // 手机端的编号
// 	Brand string `json:"brand"` // 手机端的品牌
// 	Local string `json:"local"` // 代理ip的国家iso
// }

// 任务表
// 任务可以执行3种,使用Tp(type)表示:0-打开指纹浏览器执行  1-打开手机设备执行  2-执行内置http请求
type Task struct {
	ID        int64    `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Title     string   `json:"title" gorm:"index;not null;type:varchar(32)" form:"title"`          // 任务名称
	Tp        int      `json:"type" gorm:"index;default:0" form:"type"`                            // 任务类型,分2种, 0-web端  1-手机端  2-使用golang发起http请求
	Addtime   int64    `json:"addtime" gorm:"index;default:0"`                                     // 创建时间
	Tags      []string `json:"tags" gorm:"-" form:"tags"`                                          // 设备标签
	Starttime int64    `json:"starttime" gorm:"index;default:0" form:"starttime" parse:"datetime"` // 任务开始时间
	Endtime   int64    `json:"endtime" gorm:"index;default:0" form:"endtime" parse:"datetime"`     // 任务结束时间
	Status    int      `json:"status" gorm:"type:tinyint(1);default:0;index" form:"status"`        // 任务状态, 1-可执行 0-不可执行
	Script    int64    `json:"script" gorm:"index;not null" form:"script"`                         // 任务脚本
	ScriptStr string   `json:"script_str"`                                                         // 执行的脚本字符串
	Errmsg    string   `json:"errmsg" gorm:"default:null" form:"errmsg"`                           // 错误信息
	AdminId   int64    `json:"admin_id" gorm:"index;not null"`                                     // 管理员id
	Cycle     int64    `json:"cycle" gorm:"default:0" form:"cycle"`                                // 任务周期,单位秒,0为不重复执行,大于0表示间隔多久自动重复执行
	RetryMax  int      `json:"retry_max" gorm:"default:0" form:"retry_max"`                        // 最大重试次数
	Timeout   int64    `json:"timeout" gorm:"default:0" form:"timeout"`                            // 单次超时（秒）
	// Priority      int                   `json:"priority" gorm:"default:0" form:"priority"`                          // 优先级
	// CatchUp       bool                  `json:"catch_up" gorm:"default:false" form:"catch_up"`                      // 补跑漏掉的周期
	SeNum int `json:"se_num" gorm:"default:2" form:"se_num"` // 同时执行的设备数量,0表示所有设备同时执行
	// DataSpec      string                `json:"data_spec" gorm:"default:null" form:"data_spec"` // 数据来源配置（JSON）,这种方式需要的参数
	// DataType      string                `json:"data_type" gorm:"default:null" form:"data_type"` // 数据类型标识,我用哪一种“取数方式”
	Devices  []int64        `json:"devices" gorm:"-" form:"devices"`           // 设备列表
	Params   []*TaskParam   `json:"params" gorm:"-" form:"params"`             // 参数
	DefUrl   string         `json:"def_url" form:"def_url"`                    // 默认url地址
	Headless int            `json:"headless" gorm:"type:tinyint(1);default:0"` // 0为静默(不显示窗口) 1为显示窗口
	Clients  []*TaskClients `json:"-" gorm:"-"`
	// RunnerBrowser *browserdb.Browser    `json:"-" gorm:"-"`                                // 执行的浏览器
	// RunnerPhone   *clients.Phone        `json:"-" gorm:"-"`                                // 执行的手机
	// ErrMsg   string                          `json:"err_msg" gorm:"-"` // 任务执行错误消息
	// mu       sync.Mutex                      `json:"-" gorm:"-"`       // 锁
	// isRuning bool                            `json:"-" gorm:"-"`       // 是否在执行
	// Callback func(string) error              `json:"-" gorm:"-"`       // 任务执行结果回调
	// OnError  func(error, *bs.Browser)        `json:"-" gorm:"-"`       // 任务错误结果回调
	// OnClose  func()                          `json:"-" gorm:"-"`       // 浏览器关闭回调
	// OnChange func(string, *bs.Browser) error `json:"-" gorm:"-"`       // 当浏览器地址改变回调
	// slots    chan struct{}                   `json:"-" gorm:"-"`       // 启动的协程
	// runners  map[string]*TaskClients         `json:"-" gorm:"-"`       // 任务中具体执行的设备
	// isRun atomic.Bool `json:"-" gorm:"-"` // 是否在执行中
	// runnerBrowser map[int64]*bs.Browser `json:"-" gorm:"-"` // 正在执行的bs
	// Params    string   `json:"params" gorm:"default:null" parse:"json"`                            // 脚本参数
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

// 任务对于的标签表
type TaskToTag struct {
	TaskID int64 `json:"task_id" gorm:"primaryKey"`
	TagID  int64 `json:"tag_id" gorm:"primaryKey"`
}

var dbs = db.TaskDB
var Seched *scheduler.Scheduler

func init() {
	dbs.AutoMigrate(&Task{})
	dbs.AutoMigrate(&TaskClients{})
	dbs.AutoMigrate(&TaskTag{})
	dbs.AutoMigrate(&TaskToTag{})
	dbs.AutoMigrate(&TaskRun{})
	dbs.AutoMigrate(&TaskLog{})
	dbs.AutoMigrate(&TaskParam{})

	var tsks []*Task
	dbs.Model(&Task{}).Where("status = 1").Find(&tsks)

	Seched = scheduler.New()

	for _, v := range tsks {
		go v.Listen()
	}
}

func (this *Task) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}

	if this.Title == "" {
		return fmt.Errorf("请填写任务标题")
	}

	if this.Status == 1 {
		if this.Script < 1 {
			return fmt.Errorf("请设置脚本，否则任务无法启动!")
		}
		if len(this.Devices) < 1 {
			return fmt.Errorf("请设置执行客户端，否则任务无法启动!")
		}
	}

	older := new(Task)

	defer func() {
		if older.ID > 0 {
			if older.Status != this.Status {
				if this.Status == 1 {
					this.Clients = this.GetClients()
					go this.Start()
				} else {
					go this.Stop()
				}
			}
		}

	}()

	if this.SeNum < 1 {
		this.SeNum = 2
	}
	if this.ID > 0 {
		dbs.Model(&Task{}).Where("id = ?", this.ID).First(older)
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
				// "priority":  this.Priority,
				// "catch_up":  this.CatchUp,
				"script": this.Script,
				"se_num": this.SeNum,
				// "data_spec": this.DataSpec,
				// "data_type": this.DataType,
				"def_url":  this.DefUrl,
				"headless": this.Headless,
			}).Error
	} else {
		if this.Addtime < 1 {
			this.Addtime = time.Now().Unix()
		}
		err := tx.Create(this).Error
		if err == nil {
			ws.SentBus(this.AdminId, "task", this, "")
		}
		return err
	}
}

// 获取Task
func GetTaskById(id any) *Task {
	tsk := new(Task)
	dbs.Model(&Task{}).Where("id = ?", id).First(tsk)
	return tsk
}

// 获取任务总数
func GetTotalTask(groupname string, adminid int64) int64 {
	var total int64
	md := dbs.Model(&Task{}).Where("admin_id = ?", adminid)
	if groupname != "" {
		md.Where("group_name = ?", groupname)
	}
	md.Count(&total)
	return total
}

// 获取分组的任务
func GetTasks(adminid int64, dt *db.ListFinder) ([]*Task, int64) {
	var tks []*Task
	if dt.Page < 1 {
		dt.Page = 1
	}
	if dt.Limit < 1 {
		dt.Limit = 20
	}
	md := dbs.Model(&Task{}).Where("admin_id = ?", adminid)
	if dt.Q != "" {
		qs := fmt.Sprintf("%%%s%%", dt.Q)
		md.Where("title like ?", qs)
	}

	if len(dt.Types) > 0 {
		md.Where("tp in ?", dt.Types)
	}

	if len(dt.Tags) > 0 {
		var taskids []int64
		dbs.Model(&TaskToTag{}).Select("task_id").Where("tag_id in ?", dt.Tags).Find(&taskids)
		if len(taskids) > 0 {
			md.Where("id in ?", taskids)
		}
	}

	var total int64
	md.Count(&total)

	if dt.Scol != "" && dt.By != "" {
		var byy string
		if strings.Contains(dt.By, "desc") {
			byy = "desc"
		} else {
			byy = "asc"
		}
		md.Order(fmt.Sprintf("%s %s", dt.Scol, byy))
	} else {
		md.Order("id DESC")
	}
	md.Offset((dt.Page - 1) * dt.Limit).Limit(dt.Limit).Find(&tks)

	for _, v := range tks {
		v.Devices = v.GetDevices()
		for _, zv := range v.GetTags() {
			v.Tags = append(v.Tags, zv.Name)
		}
		v.GetParams()
	}
	return tks, total
}

// 获取任务下的设备列表
func (this *Task) GetDevices() []int64 {
	var dids []int64
	dbs.Model(&TaskClients{}).Select("device_id").Where("task_id = ?", this.ID).Find(&dids)
	return dids
}

// 处理设备
func (this *Task) GenDevices() error {
	if this.ID > 0 {
		this.removeNotUsedDevices(this.Devices)
	}

	var dvs []*TaskClients
	for _, v := range this.Devices {
		dvs = append(dvs, &TaskClients{
			TaskID:     this.ID,
			DeviceType: this.Tp,
			DeviceID:   v,
		})
	}
	if len(dvs) > 0 {
		return dbs.
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&dvs).Error
	}
	return nil
}

// 删除不存在的设备
func (this *Task) removeNotUsedDevices(deviceIDs []int64) error {
	return dbs.
		Where("task_id = ?", this.ID).
		Where("device_id not in ? or device_type != ?", deviceIDs, this.Tp).
		Delete(&TaskClients{}).Error
}

// 获取任务的tags
func (this *Task) GetTags() []*TaskTag {
	var ttids []int64
	dbs.Model(&TaskToTag{}).Select("tag_id").Where("task_id = ?", this.ID).Find(&ttids)

	var tags []*TaskTag
	dbs.Model(&TaskTag{}).Where("id in ?", ttids).Find(&tags)
	return tags
}

// 通过task添加tags
func (this *Task) AddTags() error {
	tgs := AddTagsBySlice(this.Tags) // 不管三七二十一,将标签在标签表内添加一遍
	var tagIds []int64
	for _, v := range tgs {
		tagIds = append(tagIds, v.ID)
	}

	dbs.Where("task_id = ?", this.ID).Where("tag_id not in ?", tagIds).Delete(&TaskToTag{}) // 不管三七二十一,将对应表中不存在的标签id删除

	if len(tagIds) > 0 {
		tags := make([]*TaskToTag, 0, len(tagIds))
		for _, tid := range tagIds {
			tags = append(tags, &TaskToTag{
				TaskID: this.ID,
				TagID:  tid,
			})
		}

		return dbs.
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&tags).Error
	}
	return nil
}

func DeleteByID(id any) error {
	tsk := GetTaskById(id)
	if tsk != nil && tsk.ID > 0 {
		if err := dbs.Where("id = ?", id).Delete(&Task{}).Error; err != nil {
			return err
		}
		// Seched.Remove(GenTaskID(tsk.ID))
	}

	return nil
}

func (t *Task) GetRunJscript() string {
	if t.ScriptStr == "" && t.Script > 0 {
		js := jses.GetJsById(t.Script)
		if js != nil && js.ID > 0 {
			params := t.GetParams()
			mp := make(map[string]any)
			for _, v := range params {
				mp[v.CodeName] = v.Value
			}
			t.ScriptStr = js.GetContent(mp)
		}
	}
	return t.ScriptStr
}

// 清空任务
func Flush() {

}
