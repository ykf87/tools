package langs

import (
	"fmt"
	"tools/runtimes/db"
)

type Lang struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Code string `json:"code" gorm:"uniqueIndex;not null"`
	Name string `json:"name" gorm:"index"`
}

var DB = db.PRODUCTDB

func init() {
	DB.DB().AutoMigrate(&Lang{})
}

func GetLangs(req db.ListFinder) ([]*Lang, int64) {
	model := DB.DB().Model(&Lang{})

	if req.Q != "" {
		model = model.Where("code = ?", req.Q).Or("name like ?", fmt.Sprintf("%%%s%%", req.Q))
	}

	var total int64
	model.Count(&total)

	if req.Limit > 0 {
		if req.Page < 1 {
			req.Page = 1
		}
		model.Offset((req.Page - 1) * req.Limit).Limit(req.Limit)
	}
	var langs []*Lang
	model.Order("id ASC").Find(&langs)
	return langs, total
}
