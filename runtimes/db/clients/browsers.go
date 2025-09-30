package clients

import (
	"time"
	"tools/runtimes/db"

	"gorm.io/gorm"
)

type Browser struct {
	Id          int64  `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name        string `json:"name" gorm:"index;not null" form:"name"`                // 名称
	Proxy       int64  `json:"proxy" gorm:"index;default:0" form:"proxy"`             // 代理
	ProxyConfig string `json:"proxy_config" gorm:"default:null;" form:"proxy_config"` // 代理的配置项,如果设置了此值,proxy失效
	Lang        string `json:"lang" gorm:"default:null;" form:"lang"`                 // 语言
	Timezone    string `json:"timezone" gorm:"default:null;" form:"timezone"`         // 时区
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (this *Browser) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Model(&Browser{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"name":         this.Name,
				"proxy":        this.Proxy,
				"proxy_config": this.ProxyConfig,
				"lang":         this.Lang,
				"timezone":     this.Timezone,
				"updated_at":   time.Now(),
			}).Error
	} else {
		this.CreatedAt = time.Now()
		this.UpdatedAt = this.CreatedAt
		return tx.Create(this).Error
	}
}

func (this *Browser) Open(proxyUrl string, width, height int) {

}
