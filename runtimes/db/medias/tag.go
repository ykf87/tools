package medias

import "gorm.io/gorm/clause"

type MediaUserTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
}

func GetTags() []*MediaUserTag {
	var tags []*MediaUserTag
	dbs.Model(&MediaUserTag{}).Find(&tags)
	return tags
}

func GetMUTagsByNames(names []string) []*MediaUserTag {
	var tgs []*MediaUserTag
	dbs.Model(&MediaUserTag{}).Where("name in ?", names).Find(&tgs)
	return tgs
}

// 获取tag的id
func AddMUTagsBySlice(tagNames []string) []*MediaUserTag {
	ln := len(tagNames)
	if ln > 0 {
		tags := make([]*MediaUserTag, 0, ln)
		for _, name := range tagNames {
			tags = append(tags, &MediaUserTag{
				Name: name,
			})
		}

		dbs.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&tags)

		return GetMUTagsByNames(tagNames)
	}
	return nil
}
