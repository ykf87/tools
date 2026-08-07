package spider

import (
	"time"
	"tools/runtimes/db"
)

type SpiderHistory struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	JSID      int64     `json:"js_id" gorm:"index"` // 使用的js
	JSStr     string    `json:"js_str"`             // jsstr 和 jsid 只能有一个
	Url       string    `json:"url"`                // 采集地址
	Html      string    `json:"html"`               // html和url只能有一个
	Config    string    `json:"config"`             // 采集的配置
	CreatedAt time.Time `json:"created_at"`         // 采集时间
	db.BaseModel
}
