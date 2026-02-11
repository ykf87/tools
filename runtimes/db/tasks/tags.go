package tasks

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 任务标签表
type TaskTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
}

// 获取所有tags
func GetTags() []*TaskTag {
	var tgs []*TaskTag
	Dbs.DB().Model(&TaskTag{}).Find(&tgs)
	return tgs
}

func GetTagsByNames(names []string) []*TaskTag {
	var tgs []*TaskTag
	Dbs.DB().Model(&TaskTag{}).Where("name in ?", names).Find(&tgs)
	return tgs
}

// 添加标签,通过字符串数组
func AddTagsBySlice(tagNames []string) []*TaskTag {
	ln := len(tagNames)
	if ln > 0 {
		tags := make([]*TaskTag, 0, ln)
		for _, name := range tagNames {
			tags = append(tags, &TaskTag{
				Name: name,
			})
		}

		Dbs.Write(func(tx *gorm.DB) error {
			return tx.Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&tags).Error
		})
		// dbs.Clauses(clause.OnConflict{
		// 	DoNothing: true,
		// }).Create(&tags)

		return GetTagsByNames(tagNames)
	}
	return nil
}
