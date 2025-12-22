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
	Removeable int    `json:"removeable" gorm:"index;default:1"` // 1为可删除,0为不允许删除
}

var cfgMap sync.Map
var defData = map[string]*Config{
	"autojoin": &Config{
		Key:        "autojoin",
		Value:      "1",
		Remark:     "是否允许手机端扫码自动添加",
		Removeable: 0,
	},
	"taskremoveDay": &Config{
		Key:        "taskremoveDay",
		Value:      "7",
		Remark:     "任务保留天数",
		Removeable: 0,
	},
}

func init() {
	db.DB.AutoMigrate(&Config{})
	var lst []*Config
	db.DB.Model(&Config{}).Find(&lst)
	for _, c := range lst {
		cfgMap.Store(c.Key, c)
	}

	// 添加默认值
	for k, v := range defData {
		if _, ok := GetValue(k); !ok {
			if err := v.Save(nil); err == nil {
				cfgMap.Store(v.Key, v)
			}
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

// 获取某个键的值
func GetValue(k string) (string, bool) {
	if s, ok := cfgMap.Load(k); ok {
		if obj, okk := s.(*Config); okk {
			return obj.Value, true
		}
		return "", false
	}
	return "", false
}

// 获取某个键的实例
func GetObject(k string) *Config {
	if s, ok := cfgMap.Load(k); ok {
		if obj, okk := s.(*Config); okk {
			return obj
		}
		return nil
	}
	return nil
}

// 获取配置列表,可分页
func GetList(page, limit int) []*Config {
	var lst []*Config
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	db.DB.Model(&Config{}).Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&lst)
	return lst
}
