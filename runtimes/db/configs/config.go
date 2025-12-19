package configs

import (
	"sync"
	"time"
	"tools/runtimes/db"

	"gorm.io/gorm"
)

type Config struct {
	Id         int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Key        string `json:"key" gorm:"uniqueIndex;not null;type:varchar(32)"`
	Value      string `json:"value" gorm:"default:null;"`
	Remark     string `json:"remark" gorm:"default:null"`
	Addtime    int64  `json:"addtime" gorm:"index;default:0"`
	Removeable int    `json:"removeable" gorm:"index;default:0"` // 0为可删除,1为不允许删除
}

var cfgMap sync.Map
var defData = map[string]*Config{
	"autojoin": &Config{
		Key:        "autojoin",
		Value:      "1",
		Remark:     "是否允许手机端自动扫码加入",
		Removeable: 1,
	},
}

func init() {
	db.DB.AutoMigrate(&Config{})
	var lst []*Config
	db.DB.Model(&Config{}).Find(&lst)
	for _, c := range lst {
		cfgMap.Store(c.Key, c)
	}
	for k, v := range defData {
		if _, ok := GetValue(k); !ok {
			v.Save(nil)
		}
	}
}

func (this *Config) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Model(&Config{}).Where("id = ?", this.Id).
			Updates(map[string]any{
				"value":      this.Value,
				"remark":     this.Remark,
				"removeable": this.Removeable,
			}).Error
	} else {
		this.Addtime = time.Now().Unix()
		return tx.Create(this).Error
	}
}

func GetValue(k string) (string, bool) {
	if s, ok := cfgMap.Load(k); ok {
		if obj, okk := s.(*Config); okk {
			return obj.Value, true
		}
		return "", false
	}
	return "", false
}
