package clients

type Browser struct {
	Id          int64  `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name        string `json:"name" gorm:"index;not null" form:"name"`                // 名称
	Proxy       int64  `json:"proxy" gorm:"index;default:0" form:"proxy"`             // 代理
	ProxyConfig string `json:"proxy_config" gorm:"default:null;" form:"proxy_config"` // 代理的配置项,如果设置了此值,proxy失效
	Lang        string `json:"lang" gorm:"default:null;" form:"lang"`                 // 语言
}
