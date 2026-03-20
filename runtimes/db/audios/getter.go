package audios

import (
	"fmt"
	"tools/runtimes/db"
)

func GetList(s *db.ListFinder) (total int64, lists []*Audio) {
	page := 1
	limit := 20
	if s.Page > 0 {
		page = s.Page
	}
	if s.Limit > 0 {
		limit = s.Limit
	}

	model := Dbs.DB().Debug().Model(&Audio{}).Where("removed = 0")
	if s.Q != "" {
		model = model.Where("title like ?", fmt.Sprintf("%%%s%%", s.Q))
	}

	if s.Filters != nil {
		for name, val := range s.Filters {
			switch name {
			case "tags":
				var tagIdArr []int64
				if tags, ok := val.([]any); ok {
					for _, tagID := range tags {
						if idf, ok := tagID.(float64); ok {
							tagIdArr = append(tagIdArr, int64(idf))
						}
					}
				}
				fmt.Println(tagIdArr, len(tagIdArr))
				if len(tagIdArr) > 0 {
					// model = model.Preload("Tags", func(db *gorm.DB) *gorm.DB {
					// 	return db.Where("id IN ?", tagIdArr)
					// })
					model = model.Joins("JOIN audio_tag_relations atr ON atr.audio_id = audios.id").
						Where("atr.audio_tag_id IN ?", tagIdArr).Distinct()
				}
			}
		}
	}

	model.Count(&total)

	model.Preload("Tags").Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&lists)

	return
}
