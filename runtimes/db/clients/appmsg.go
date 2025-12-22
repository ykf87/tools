// 给app发送的任务什么的存储到数据库
package clients

import (
	"errors"
	"strconv"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/db/configs"

	"gorm.io/gorm"
)

// app 任务,任务一般存在内存,如果app未连接,则下次连接从数据库获取,已完成的任务保留n天后删除
// 任务添加时并不执行启动
type AppTask struct {
	Id        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Type      string `json:"type" gorm:"index;not null"`       // 任务类型
	Data      string `json:"data" gorm:"default:null"`         // 任务数据
	Addtime   int64  `json:"addtime" gorm:"index;default:0"`   // 添加时间
	Starttime int64  `json:"starttime" gorm:"index;default:0"` // 任务开始时间
	Endtime   int64  `json:"endtime" gorm:"index;default:0"`   // 任务结束或完结时间,包括成功和失败
	Status    int    `json:"status" gorm:"index;default:0"`    // 任务状态
	ErrMsg    string `json:"err_msg" gorm:"default:null"`      // 错误信息
	DeviceId  string `json:"device_id" gorm:"index;not null"`  // 设备唯一id
	AdminId   int64  `json:"admin_id" gorm:"index;default:0"`  // 管理员id
	Cycle     int64  `json:"cycle" gorm:"default:0"`           // 任务周期,单位秒,0为不重复执行,大于0表示间隔多久自动重复执行
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
