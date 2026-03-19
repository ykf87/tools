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

	model := Dbs.DB().Model(&Audio{}).Where("removed = 0")
	if s.Q != "" {
		model = model.Where("title like ?", fmt.Sprintf("%%%s%%", s.Q))
	}

	model.Count(&total)
	model.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&lists)

	return
}
