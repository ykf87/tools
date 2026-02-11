package information

// 消息

import (
	"tools/runtimes/db"
	"tools/runtimes/eventbus"
)

type Information struct {
	Id       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Tab      string `json:"tab" gorm:"index;type:varchar(16);default:null"` // tab标签
	Title    string `json:"title" gorm:"not null"`                          // 通知标题
	Content  string `json:"content" gorm:"default:null"`                    // 通知内容
	Jump     string `json:"jump" gorm:"default:null"`                       // 跳转连接
	JumpTxt  string `json:"jump_txt" gorm:"default:null"`                   // 跳转文本
	JumpIcon string `json:"jump_icon"`                                      // 跳转按钮的图标
	Readtime int64  `json:"readtime" gorm:"index;default:0"`                // 查看时间
	AdminId  int64  `json:"admin_id" gorm:"index;default:0"`                // 发给哪个用户的,0为本地所有用户
	Addtime  int64  `json:"addtime" gorm:"index;default:0"`                 // 添加时间
	db.BaseModel
}

func init() {
	db.DB.DB().AutoMigrate(&Information{})
}

// func (this *Information) Save(tx *db.SQLiteWriter) error {
// 	if tx == nil {
// 		tx = db.DB
// 	}
// 	var err error
// 	if this.Id > 0 {
// 		err = tx.Write(func(txx *gorm.DB) error {
// 			return txx.Model(&Information{}).Where("id = ?", this.Id).
// 				Updates(map[string]any{
// 					"title":     this.Title,
// 					"content":   this.Content,
// 					"jump":      this.Jump,
// 					"jump_txt":  this.JumpTxt,
// 					"jump_icon": this.JumpIcon,
// 					"readtime":  this.Readtime,
// 					"admin_id":  this.AdminId,
// 				}).Error
// 		})
// 	} else {
// 		this.Addtime = time.Now().Unix()
// 		err = tx.Write(func(txx *gorm.DB) error {
// 			return txx.Create(this).Error
// 		})
// 		if err == nil {
// 			go ws.SentBus(this.AdminId, "information", this, "")
// 		}
// 	}
// 	return err
// }

func (t *Information) SendWs() {
	eventbus.Bus.Publish("information", t)
}

// 获取所有的tab
type TabStru struct {
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
	TabName string         `json:"tab_name"`
	Loading bool           `json:"loading"`
	More    bool           `json:"more"`
	List    []*Information `json:"list"`
	Key     string         `json:"key"`
	Total   int64          `json:"total"`
	NotRead int64          `json:"not_read"`
}

const DEFLIMIT = 20

func GetInforTabs(uid int64) []TabStru {
	var tabs []string
	db.DB.DB().Model(&Information{}).Select("tab").Where("admin_id = 0 or admin_id = ?", uid).Group("tab").Find(&tabs)

	var total int64
	var notread int64
	db.DB.DB().Model(&Information{}).Where("admin_id = 0 or admin_id = ?", uid).Count(&total)
	db.DB.DB().Model(&Information{}).Where("admin_id = 0 or admin_id = ?", uid).Where("readtime = 0").Count(&notread)
	tabstru := []TabStru{
		TabStru{
			Page:    1,
			Limit:   DEFLIMIT,
			TabName: "全部",
			Key:     "",
			More:    true,
			Total:   total,
			NotRead: notread,
		},
	}
	for _, v := range tabs {
		var ttotal int64
		var tnread int64
		db.DB.DB().Model(&Information{}).Where("admin_id = 0 or admin_id = ?", uid).Where("tab = ?", v).Count(&ttotal)
		db.DB.DB().Model(&Information{}).Where("admin_id = 0 or admin_id = ?", uid).Where("readtime = 0").Where("tab = ?", v).Count(&tnread)
		tabstru = append(tabstru, TabStru{
			Page:    1,
			Limit:   DEFLIMIT,
			TabName: v,
			Key:     v,
			More:    true,
			Total:   ttotal,
			NotRead: tnread,
		})
	}
	return tabstru
}

// 获取消息总数
func GetInfomationsTotal(adminid int64, tab string) int64 {
	var total int64
	db.DB.DB().Model(&Information{}).
		Where("admin_id = ? or admin_id = 0", adminid).Where("tab = ?", tab).Count(&total)
	return total
}

// 分页获取通知消息
func GetInfomation(page, limit int, adminid int64, tab string) []*Information {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	var nfs []*Information
	model := db.DB.DB().Model(&Information{}).
		Where("admin_id = ? or admin_id = 0", adminid)
	if tab != "" {
		model.Where("tab = ?", tab)
	}
	model.Order("readtime asc, id DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&nfs)
	return nfs
}

// 获取未读消息
func GetNotRead(uid int64, tab string, page, limit int) []*Information {
	model := db.DB.DB().Model(&Information{})
	if uid > 0 {
		model.Where("admin_id = 0 or admin_id = ?", uid)
	} else {
		model.Where("admin_id = 0")
	}
	if tab != "" {
		model.Where("tab = ?", tab)
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = DEFLIMIT
	}

	var list []*Information
	model.Offset((page - 1) * limit).Limit(limit).Order("id desc").Find(&list)
	return list
}
