package spider

import (
	"time"
	"tools/runtimes/db"
)

type SpiderJs struct { // 如果有修改字段,需要更新Save方法
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name      string    `json:"name" gorm:"index;not null" form:"id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	Content   string    `json:"content"`
	db.BaseModel
}

func init() {
	db.DB.DB().AutoMigrate(&SpiderJs{})
}

func GetSpiderJses() []*SpiderJs {
	var jses []*SpiderJs
	db.DB.DB().Order("created_at desc").Find(&jses)
	return jses
}
