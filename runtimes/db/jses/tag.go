package jses

import "tools/runtimes/db"

type JsTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"type:varchar(60);index"`
}

type JsToTag struct {
	JsID  int64 `json:"js_id" gorm:"primaryKey;not null"`
	TagID int64 `json:"tag_id" gorm:"primaryKey;not null"`
}

func init() {
	db.DB.AutoMigrate(&JsTag{})
	db.DB.AutoMigrate(&JsToTag{})
}

// 获取脚本下的tags
func (this *Js) GetTags() []*JsTag {
	var ttids []int64
	db.DB.Model(&JsToTag{}).Select("tag_id").Where("js_id = ?", this.ID).Find(&ttids)

	var tags []*JsTag
	db.DB.Model(&JsTag{}).Where("id in ?", ttids).Find(&tags)
	return tags
}

// 获取脚本的标签列表
func GetTags() []*JsTag {
	var tgs []*JsTag

	db.DB.Model(&JsTag{}).Find(&tgs)
	return tgs
}
