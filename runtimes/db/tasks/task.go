package tasks

import (
	"time"
	"tools/runtimes/db"
	"tools/runtimes/listens/ws"

	"gorm.io/gorm"
)

type Task struct {
	Id        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Starttime int64  `json:"starttime" gorm:"index;not null"`                       // 任务开始时间
	Endtime   int64  `json:"endtime" gorm:"index;default:0"`                        // 任务结束时间
	Status    int    `json:"status" gorm:"type:tinyint(1);default:0;index"`         // 任务状态,1成功,0之中,-1失败
	Errmsg    string `json:"errmsg" gorm:"default:null"`                            // 错误信息
	AdminId   int64  `json:"admin_id" gorm:"index;not null"`                        // 管理员id
	Group     string `json:"group_name" gorm:"type:varchar(32);index;default:null"` // 任务类型
}

func init() {
	db.TaskDB.AutoMigrate(&Task{})
}

func (this *Task) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.TaskDB
	}
	if this.Id > 0 {
		return tx.Model(&Task{}).Where("id = ?", this.Id).
			Updates(map[string]any{
				"starttime": this.Starttime,
				"endtime":   this.Endtime,
				"status":    this.Status,
				"errmsg":    this.Errmsg,
				"admin_id":  this.AdminId,
				"group":     this.Group,
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

// 获取所有的group name
func GetGroup(adminid int64) []string {
	var gps []string
	db.TaskDB.Model(&Task{}).Select("group_name").Where("admin_id = ?", adminid).Group("group_name").Find(&gps)
	return gps
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
func GetTasks(page, limit int, adminid int64, groupname string) []*Task {
	var tks []*Task
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	md := db.TaskDB.Model(&Task{}).Where("admin_id = ?", adminid)
	if groupname != "" {
		md.Where("group_name = ?", groupname)
	}

	md.Order("starttime DESC").Offset((page - 1) * limit).Limit(limit).Find(&tks)
	return tks
}
