// 给app发送的任务什么的存储到数据库
package clients

import (
	"context"
	"errors"
	"strconv"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/configs"

	"gorm.io/gorm"
)

// app 任务,任务一般存在内存,如果app未连接,则下次连接从数据库获取,已完成的任务保留n天后删除
// 任务添加时并不执行启动,需要执行 *AppTask.Run() 启动
type AppTask struct {
	Id        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Type      string `json:"type" gorm:"index;not null"`       // 任务类型
	Data      string `json:"data" gorm:"default:null"`         // 任务数据
	Addtime   int64  `json:"addtime" gorm:"index;default:0"`   // 添加时间
	Status    int    `json:"status" gorm:"index;default:0"`    // 任务状态 0:待启动 1:执行中
	ErrMsg    string `json:"err_msg" gorm:"default:null"`      // 错误信息
	DeviceId  string `json:"device_id" gorm:"index;not null"`  // 设备唯一id
	AdminId   int64  `json:"admin_id" gorm:"index;default:0"`  // 管理员id
	Starttime int64  `json:"starttime" gorm:"index;default:0"` // 任务允许的开始时间,也就是可以设置在某个时间段执行
	Endtime   int64  `json:"endtime" gorm:"index;default:0"`   // 任务关闭时间,也就是大于这个时间不执行
	Cycle     int64  `json:"cycle" gorm:"default:0"`           // 任务周期,单位秒,0为不重复执行,大于0表示间隔多久自动重复执行
}

// 由于要兼顾任务可定时执行和周期性任务,因此需要额外增加一张表存储任务执行情况
type AppTaskMsg struct {
	RunId     int64  `json:"run_id" gorm:"primaryKey;autoIncrement"`
	TaskId    int64  `json:"task_id" gorm:"index"`
	RunStatus int    `json:"run_status" gorm:"type:tinyint(1);index;default:0"` // 运行结果状态
	Msg       string `json:"msg" gorm:"default:null"`                           // 执行结果
	Exectime  int64  `json:"exectime" gorm:"index;default:0"`                   // 任务开始执行时间
	Donetime  int64  `json:"donetime" gorm:"index;default:0"`                   // 任务结束或完结时间,包括成功和失败
}

var dbs = db.AppTask

func init() {
	dbs.AutoMigrate(&AppTask{})
	rmvDay := 7
	if v, ok := configs.GetValue("taskremoveDay"); ok {
		if vv, err := strconv.Atoi(v); err == nil {
			rmvDay = vv
		}
	}
	rmvDay64 := int64(rmvDay * 86400)
	go func() { // 定时清除过期任务
		for {
			dbs.Where("addtime < ?", (time.Now().Unix() - rmvDay64)).Where("status != 0").Delete(&AppTask{})
			time.Sleep(time.Hour * 24)
		}
	}()
}

func (this *AppTask) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.AppTask
	}
	if this.Id > 0 {
		return tx.Model(&AppTask{}).Where("id = ?", this.Id).
			Updates(map[string]any{
				"type":      this.Type,
				"data":      this.Data,
				"starttime": this.Starttime,
				"endtime":   this.Endtime,
				"status":    this.Status,
				"err_msg":   this.ErrMsg,
				// "device_id": this.DeviceId,
				"cycle": this.Cycle,
			}).Error
	} else {
		if this.Addtime < 1 {
			this.Addtime = time.Now().Unix()
		}
		// if this.AdminId < 1 {
		// 	return errors.New("缺少adminid")
		// }
		if this.DeviceId == "" {
			return errors.New("缺少设备id")
		}
		return tx.Create(this).Error
	}
}

// 运行任务
func (this *AppTask) Run() {
	this.Status = 1
	if this.Save(nil) == nil {
		if msg, err := config.Json.Marshal(this); err == nil {
			Hubs.SentClient(this.DeviceId, msg)
		}
	}
}

// 1️⃣ 判断是否允许执行
func (this *AppTask) CanRunTask() bool {
	now := time.Now().Unix()

	if this.Starttime > 0 && now < this.Starttime {
		return false
	}
	if this.Endtime > 0 && now > this.Endtime {
		return false
	}
	return true
}

// 2️⃣ 是否存在“正在执行”的记录
func (this *AppTask) GetRunningMsg() (*AppTaskMsg, error) {
	var msg AppTaskMsg
	err := dbs.
		Where("task_id = ? AND run_status = 0", this.Id).
		Order("run_id DESC").
		First(&msg).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &msg, err
}

// 3️⃣ 创建一次执行（安全）
func (this *AppTask) CreateRun() (*AppTaskMsg, error) {
	msg := &AppTaskMsg{
		TaskId:    this.Id,
		RunStatus: 0,
		Exectime:  time.Now().Unix(),
	}

	err := dbs.Create(msg).Error
	if err != nil {
		return nil, err
	}
	return msg, nil
}

/*
五、重点来了：任务循环执行核心方法
✅ 满足你提出的所有条件
✅ 上一次不完成，下一次绝不开始
✅ 支持 cycle / 非 cycle
✅ 可被 WS / API 直接调用
*/
func (task *AppTask) RunTaskLoop(
	ctx context.Context,
	execute func(run *AppTaskMsg) error,
) {
	for {
		// 1. 是否还允许执行
		if !task.CanRunTask() {
			return
		}

		// 2. 是否有正在执行的
		running, err := task.GetRunningMsg()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		if running != nil {
			// 等待当前执行完成
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
					var check AppTaskMsg
					err := dbs.First(&check, running.RunId).Error
					if err != nil {
						continue
					}
					if check.RunStatus != 0 {
						goto NEXT
					}
				}
			}
		}

	NEXT:
		// 3. 创建新执行
		run, err := task.CreateRun()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		// 4. 执行（同步）
		err = execute(run)

		// 5. 更新执行结果
		updates := map[string]interface{}{
			"donetime": time.Now().Unix(),
		}
		if err != nil {
			updates["run_status"] = 2
		} else {
			updates["run_status"] = 1
		}
		dbs.Model(&AppTaskMsg{}).
			Where("run_id = ?", run.RunId).
			Updates(updates)

		// 6. 是否继续循环
		if task.Cycle <= 0 {
			return
		}

		// 7. 等待 cycle
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(task.Cycle) * time.Second):
		}
	}
}

// 获取需要执行的任务
func GetTask(deviceid string) []*AppTask {
	var list []*AppTask
	dbs.Model(&AppTask{}).Where("status = 1 and device_id = ? and starttime = 0", deviceid).Find(&list)
	return list
}
