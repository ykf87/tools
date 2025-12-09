package information

// 消息

import (
	"time"
	"tools/runtimes/db"
	"tools/runtimes/listens/ws"

	"gorm.io/gorm"
)

type Information struct {
	Id       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Title    string `json:"title" gorm:"not null"`           // 通知标题
	Content  string `json:"content" gorm:"default:null"`     // 通知内容
	Jump     string `json:"jump" gorm:"default:null"`        // 跳转连接
	JumpTxt  string `json:"jump_txt" gorm:"default:null"`    // 跳转文本
	JumpIcon string `json:"jump_icon"`                       // 跳转按钮的图标
	Readtime int64  `json:"readtime" gorm:"index;default:0"` // 查看时间
	AdminId  int64  `json:"admin_id" gorm:"index;default:0"` // 发给哪个用户的,0为本地所有用户
	Addtime  int64  `json:"addtime" gorm:"index;default:0"`  // 添加时间
}

func init() {
	db.DB.AutoMigrate(&Information{})
}

func (this *Information) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	var err error
	if this.Id > 0 {
		err = tx.Model(&Information{}).Where("id = ?", this.Id).
			Updates(map[string]any{
				"title":     this.Title,
				"content":   this.Content,
				"jump":      this.Jump,
				"jump_txt":  this.JumpTxt,
				"jump_icon": this.JumpIcon,
				"readtime":  this.Readtime,
				"admin_id":  this.AdminId,
			}).Error
	} else {
		this.Addtime = time.Now().Unix()
		err = tx.Create(this).Error
		if err == nil {
			go ws.SentBus(this.AdminId, "information", this, "")
		}
	}
	return err
}

// 获取通知总数
func GetInfomationsTotal(adminid int64) int64 {
	var total int64
	db.DB.Model(&Information{}).
		Where("admin_id = ? or admin_id = 0", adminid).Count(&total)
	return total
}

// 分页获取通知消息
func GetInfomation(page, limit int, adminid int64) []*Information {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	var nfs []*Information
	db.DB.Model(&Information{}).
		Where("admin_id = ? or admin_id = 0", adminid).
		Order("readtime asc, id DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&nfs)
	return nfs
}
