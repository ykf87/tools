package clients

import (
	"errors"
	"fmt"
	"tools/runtimes/db"
	"tools/runtimes/eventbus"
	"tools/runtimes/ws"

	"gorm.io/gorm"
)

type Phone struct {
	Id          int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string   `json:"name" gorm:"index;default:null;type:varchar(32)"`       // 自己备注的名称
	DeviceId    string   `json:"device_id" gorm:"index;type:varchar(64);not null"`      // 唯一的设备id
	Num         int64    `json:"num" gorm:"default:0;index"`                            // 设备编号
	AdminId     int64    `json:"admin_id" gorm:"index;not null;"`                       // 管理员编号
	Os          string   `json:"os" gorm:"index;default:null;type:varchar(32)"`         // 手机系统
	Brand       string   `json:"brand" gorm:"index;default:null;type:varchar(32)"`      // 手机品牌
	Version     string   `json:"version" gorm:"index;default:null;type:varchar(32)"`    // 系统的版本
	Addtime     int64    `json:"addtime" gorm:"index;default:0" update:"false"`         // 添加时间
	Conntime    int64    `json:"conntime" gorm:"index;default:0"`                       // 上一次连接的时间
	Proxy       int64    `json:"proxy" gorm:"index;default:0" form:"proxy"`             // 代理
	ProxyName   string   `json:"proxy_name" gorm:"index;default:null" form:"-"`         // 代理的名称,只有设置了代理id才有
	ProxyConfig string   `json:"proxy_config" gorm:"default:null;" form:"proxy_config"` // 代理的配置项,如果设置了此值,proxy失效
	Local       string   `json:"local" gorm:"default:null" form:"local"`                // 所在地区
	Lang        string   `json:"lang" gorm:"default:null;" form:"lang"`                 // 语言
	Timezone    string   `json:"timezone" gorm:"default:null;" form:"timezone"`         // 时区
	Ip          string   `json:"ip" gorm:"default:null;"`                               // ip地址,设置了代理才有
	Tags        []string `json:"tags" gorm:"-" form:"tags"`                             // 标签
	Connected   bool     `json:"connected" gorm:"-" parse:"-"`                          // 是否连接标识
	Conn        *ws.Conn `json:"-" gorm:"-" parse:"-"`                                  // 连接句柄
	Status      int      `json:"status" gorm:"index;default:0;type:tinyint(1)"`         // 设备状态
	CloseMsg    string   `json:"close_msg" gorm:"default:null"`                         // 异常提示
	Infos       string   `json:"infos" gorm:"default:null" parse:"-"`                   // 存储手机端的信息,json格式
	db.BaseModel
}

var MaxPhoneNum int64 = 2 // 最大的手机设备连接数量

type PhoneTag struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"index"`
	db.BaseModel
}

type PhoneToTag struct {
	PhoneId int64 `json:"phone_id" gorm:"primaryKey;not null"`
	TagId   int64 `json:"tag_id" gorm:"primaryKey;not null"`
	db.BaseModel
}

func init() {
	db.DB.DB().AutoMigrate(&Phone{})
	db.DB.DB().AutoMigrate(&PhoneTag{})
	db.DB.DB().AutoMigrate(&PhoneToTag{})

}

// func (t *Phone) Save(tx *db.SQLiteWriter) error {
// 	if tx == nil {
// 		tx = db.DB
// 	}
// 	var err error
// 	if t.Id > 0 {
// 		err = tx.Write(func(txx *gorm.DB) error {
// 			return txx.Model(&Phone{}).Where("id = ?", t.Id).Updates(map[string]any{
// 				"name":         t.Name,
// 				"os":           t.Os,
// 				"brand":        t.Brand,
// 				"num":          t.Num,
// 				"admin_id":     t.AdminId,
// 				"version":      t.Version,
// 				"conntime":     t.Conntime,
// 				"proxy":        t.Proxy,
// 				"proxy_name":   t.ProxyName,
// 				"proxy_config": t.ProxyConfig,
// 				"local":        t.Local,
// 				"lang":         t.Lang,
// 				"timezone":     t.Timezone,
// 				"ip":           t.Ip,
// 				"status":       t.Status,
// 				"infos":        t.Infos,
// 			}).Error
// 		})

// 		if t.Status != 1 {
// 			eventbus.Bus.Publish("phone-status-change", t)
// 		}
// 	} else {
// 		auto, ok := configs.GetValue("autojoin")
// 		if !ok || auto != "1" {
// 			return errors.New("服务端关闭自动连接,请在后台打开")
// 		}

// 		var total int64
// 		tx.DB().Model(&Phone{}).Where("status = 1").Count(&total)
// 		if total >= MaxPhoneNum {
// 			errmsg := fmt.Sprintf("设备数量超出限制: %d", MaxPhoneNum)
// 			return errors.New(errmsg)
// 		}
// 		if t.Addtime < 1 {
// 			t.Addtime = time.Now().Unix()
// 		}
// 		err = tx.Write(func(txx *gorm.DB) error {
// 			return txx.Create(t).Error
// 		})
// 	}
// 	return err
// }

// 标签操作开始-----------------------------------
// 保存标签
// func (this *PhoneTag) Save(tx *db.SQLiteWriter) error {
// 	if tx == nil {
// 		tx = db.DB
// 	}
// 	if this.Id > 0 {
// 		return tx.Write(func(txx *gorm.DB) error {
// 			return txx.Model(&PhoneTag{}).Where("id = ?", this.Id).
// 				Updates(map[string]any{
// 					"name": this.Name,
// 				}).Error
// 		})
// 	} else {
// 		return tx.Write(func(txx *gorm.DB) error {
// 			return txx.Create(this).Error
// 		})
// 	}
// }

// 删除标签
func (this *PhoneTag) Remove(tx *db.SQLiteWriter) error {
	if tx == nil {
		tx = db.DB
	}
	if this != nil && this.Id > 0 {
		err := tx.Write(func(txx *gorm.DB) error {
			return txx.Where("id = ?", this.Id).Delete(&PhoneTag{}).Error
		})
		if err != nil {
			return err
		}
		return tx.Write(func(txx *gorm.DB) error {
			return txx.Where("tag_id = ?", this.Id).Delete(&PhoneToTag{}).Error
		})
	}
	return nil
}

// 通过id获取标签
func GetPhoneTagsById(id any) *PhoneTag {
	tg := new(PhoneTag)
	db.DB.DB().Model(&PhoneTag{}).Where("id = ?", id).First(tg)
	return tg
}

// 通过标签名称获取对应的数组
func GetPhoneTagsByNames(names []string, tx *db.SQLiteWriter) map[string]int64 {
	if tx == nil {
		tx = db.DB
	}
	var tgs []*PhoneTag
	tx.DB().Model(&PhoneTag{}).Where("name in ?", names).Find(&tgs)
	mp := make(map[string]int64)
	for _, v := range tgs {
		mp[v.Name] = v.Id
	}

	var addn []*PhoneTag
	for _, v := range names {
		if _, ok := mp[v]; !ok {
			addn = append(addn, &PhoneTag{Name: v})
		}
	}
	if len(addn) > 0 {
		tx.Write(func(txx *gorm.DB) error {
			return txx.Create(&addn).Error
		})
	}
	for _, v := range addn {
		mp[v.Name] = v.Id
	}

	return mp
}

// 通过id获取对应的数组
func GetPhoneTagsByIds(ids []int64) map[int64]string {
	var tgs []*PhoneTag
	db.DB.DB().Model(&PhoneTag{}).Where("id in ?", ids).Find(&tgs)
	mp := make(map[int64]string)
	for _, v := range tgs {
		mp[v.Id] = v.Name
	}
	return mp
}

// 获取标签列表
func GetPhoneTags() []*PhoneTag {
	var tgs []*PhoneTag
	db.DB.DB().Model(&PhoneTag{}).Find(&tgs)
	return tgs
}

// tag标签结束----------------------------------------
//
// 通过id获取手机设备
func GetPhoneById(id any) (*Phone, error) {
	b := new(Phone)
	err := db.DB.DB().Model(&Phone{}).Where("id = ?", id).First(b).Error
	if err != nil {
		return nil, err
	}
	return b, nil
}

// 删除某个手机设备下的tag
func (this *Phone) RemovePhoneTags(tx *db.SQLiteWriter) error {
	if tx == nil {
		tx = db.DB
	}
	return tx.Write(func(txx *gorm.DB) error {
		return txx.Where("phone_id = ?", this.Id).Delete(&PhoneToTag{}).Error
	})
}

// 使用当前的tag标签完全替换已有标签
// 使用此方法会清空已有的tag
func (this *Phone) CoverTgs(tagsName []string, tx *db.SQLiteWriter) error {
	if tx == nil {
		tx = db.DB
	}

	if err := this.RemovePhoneTags(tx); err != nil {
		return err
	}

	mp := GetPhoneTagsByNames(tagsName, tx)

	var ntag []PhoneToTag
	for _, tagg := range tagsName {
		if tid, ok := mp[tagg]; ok {
			ntag = append(ntag, PhoneToTag{PhoneId: this.Id, TagId: tid})
		}
	}

	if len(ntag) > 0 {
		if err := tx.Write(func(txx *gorm.DB) error {
			return txx.Create(ntag).Error
		}); err != nil {
			return err
		}
	}

	return nil
}

// 设置手机设备列表的tag,调用时使用,自动将phone列表加上tag
func SetPhoneTags(pcs []*Phone) {
	if len(pcs) < 1 {
		return
	}
	var ids []int64
	for _, v := range pcs {
		ids = append(ids, v.Id)
	}

	var pxtgs []*PhoneToTag
	db.DB.DB().Model(&PhoneToTag{}).Where("phone_id in ?", ids).Find(&pxtgs)

	var tagids []int64
	pcMap := make(map[int64][]int64)
	for _, v := range pxtgs {
		tagids = append(tagids, v.TagId)
		pcMap[v.PhoneId] = append(pcMap[v.PhoneId], v.TagId)
	}

	ttggs := GetPhoneTagsByIds(tagids)

	for _, v := range pcs {
		if ids, ok := pcMap[v.Id]; ok {
			for _, tid := range ids {
				if tgname, ok := ttggs[tid]; ok {
					v.Tags = append(v.Tags, tgname)
				}
			}
		}
	}
}

func (this *Phone) Delete() error {
	if err := db.DB.Write(func(tx *gorm.DB) error {
		return tx.Where("id = ?", this.Id).Delete(&Phone{}).Error
	}); err != nil {
		return err
	}
	db.DB.Write(func(tx *gorm.DB) error {
		return tx.Where("phone_id = ?", this.Id).Delete(&PhoneToTag{}).Error
	})
	// if bbs, ok := browser.Running.Load(this.Id); ok {
	// 	if bs, ok := bbs.(*browser.User); ok {
	// 		return bs.Close()
	// 	}
	// }

	eventbus.Bus.Publish("phone-delete", this)
	return nil
}

// 通过deviceid获取设备
func GetPhoneByDeviceId(deviceId string) (*Phone, error) {
	b := new(Phone)
	err := db.DB.DB().Model(&Phone{}).Where("device_id = ?", deviceId).First(b).Error
	if err != nil {
		return nil, err
	}
	if b.Id < 1 {
		return nil, errors.New("no phone")
	}

	db.DB.DB().Select("phone_tags.name").Model(&PhoneTag{}).
		Joins("right join phone_to_tags as ptt on phone_tags.id = ptt.tag_id").
		Where("ptt.phone_id = ?", b.Id).Find(&b.Tags)
	return b, nil
}

// 获取客户端表的总数
func PhoneTotal() int64 {
	var total int64
	db.DB.DB().Model(&Phone{}).Count(&total)
	return total
}

// 获取Phone的tags 名称

// 获取手机别聊
func GetAllPhones(page, limit int, query string) []*Phone {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	model := db.DB.DB().Model(&Phone{}).Where("status = 1")
	if query != "" {
		qs := fmt.Sprintf("%%%s%%", query)
		model.Where("name like ?", qs)
	}

	var ls []*Phone
	model.Offset((page - 1) * limit).Limit(limit).Order("id desc").Find(&ls)
	return ls
}

// 通过id列表获取手机列表
func GetPhonesByIds(ids []int64) []*Phone {
	var bs []*Phone
	db.DB.DB().Model(&Phone{}).Where("id in ?", ids).Find(&bs)
	return bs
}
