package clients

import (
	"errors"
	"fmt"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/ws"

	"gorm.io/gorm"
)

type Phone struct {
	Id          int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string   `json:"name" gorm:"index;not null;type:varchar(32)"`           // 自己备注的名称
	Os          string   `json:"os" gorm:"index;default:null;type:varchar(32)"`         // 手机系统
	Brand       string   `json:"brand" gorm:"index;default:null'type:varchar(32)"`      // 手机品牌
	Version     string   `json:"varsion" gorm:"index;default:null;type:varchar(32)"`    // 系统的版本
	Addtime     int64    `json:"addtime" gorm:"index;default:0"`                        // 添加时间
	Conntime    int64    `json:"conntime" gorm:"index;default:0"`                       // 上一次连接的时间
	Proxy       int64    `json:"proxy" gorm:"index;default:0" form:"proxy"`             // 代理
	ProxyName   string   `json:"proxy_name" gorm:"index;default:null" form:"-"`         // 代理的名称,只有设置了代理id才有
	ProxyConfig string   `json:"proxy_config" gorm:"default:null;" form:"proxy_config"` // 代理的配置项,如果设置了此值,proxy失效
	Local       string   `json:"local" gorm:"default:null" form:"local"`                // 所在地区
	Lang        string   `json:"lang" gorm:"default:null;" form:"lang"`                 // 语言
	Timezone    string   `json:"timezone" gorm:"default:null;" form:"timezone"`         // 时区
	Ip          string   `json:"ip" gorm:"default:null;"`                               // ip地址,设置了代理才有
	Tags        []string `json:"tags" gorm:"-" form:"tags"`                             // 标签
	Connected   bool     `json:"connected" gorm:"-"`                                    // 是否连接标识
	Conn        *ws.Conn `json:"-" gorm:"-"`                                            // 连接句柄
}

var MaxPhoneNum int64 = 2 // 最大的手机设备连接数量

type PhoneTag struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"index"`
}

type PhoneToTag struct {
	PhoneId int64 `json:"phone_id" gorm:"primaryKey;not null"`
	TagId   int64 `json:"tag_id" gorm:"primaryKey;not null"`
}

func init() {
	db.DB.AutoMigrate(&Phone{})
	db.DB.AutoMigrate(&PhoneTag{})
	db.DB.AutoMigrate(&PhoneToTag{})
}

func (t *Phone) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	var err error
	if t.Id > 0 {
		err = tx.Model(&Phone{}).Where("id = ?", t.Id).Updates(map[string]any{
			"name":         t.Name,
			"os":           t.Os,
			"brand":        t.Brand,
			"varsion":      t.Version,
			"conntime":     t.Conntime,
			"proxy":        t.Proxy,
			"proxy_name":   t.ProxyName,
			"proxy_config": t.ProxyConfig,
			"local":        t.Local,
			"lang":         t.Lang,
			"timezone":     t.Timezone,
			"ip":           t.Ip,
		}).Error
	} else {
		var total int64
		tx.Model(&Phone{}).Count(&total)
		if total >= MaxPhoneNum {
			errmsg := fmt.Sprintf("设备数量超出限制: %d", MaxPhoneNum)
			return errors.New(errmsg)
		}
		if t.Addtime < 1 {
			t.Addtime = time.Now().Unix()
		}
		err = tx.Create(t).Error
	}
	return err
}
