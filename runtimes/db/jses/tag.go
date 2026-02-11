package jses

import (
	"tools/runtimes/db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JsTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"type:varchar(60);uniqueIndex"`
}

type JsToTag struct {
	JsID  int64 `json:"js_id" gorm:"primaryKey;not null"`
	TagID int64 `json:"tag_id" gorm:"primaryKey;not null"`
}

func init() {
	db.DB.DB().AutoMigrate(&JsTag{})
	db.DB.DB().AutoMigrate(&JsToTag{})
}

// 获取脚本下的tags
func (this *Js) GetTags() []*JsTag {
	var ttids []int64
	db.DB.DB().Model(&JsToTag{}).Select("tag_id").Where("js_id = ?", this.ID).Find(&ttids)

	var tags []*JsTag
	db.DB.DB().Model(&JsTag{}).Where("id in ?", ttids).Find(&tags)
	return tags
}

// 获取脚本的标签列表
func GetTags() []*JsTag {
	var tgs []*JsTag

	db.DB.DB().Model(&JsTag{}).Find(&tgs)
	return tgs
}

func GetTagsByNames(names []string) []*JsTag {
	var tgs []*JsTag
	db.DB.DB().Model(&JsTag{}).Where("name in ?", names).Find(&tgs)
	return tgs
}

// 添加标签,通过字符串数组
func AddTagsBySlice(tagNames []string) []*JsTag {
	ln := len(tagNames)
	if ln > 0 {
		tags := make([]*JsTag, 0, ln)
		for _, name := range tagNames {
			tags = append(tags, &JsTag{
				Name: name,
			})
		}

		db.DB.Write(func(tx *gorm.DB) error {
			return tx.Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&tags).Error
		})

		return GetTagsByNames(tagNames)
	}
	return nil
}
