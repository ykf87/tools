package products

import (
	"strings"
	"time"
)

type Cate struct {
	ID        int64       `json:"id" gorm:"primaryKey;autoIncrement"`
	Parent    int64       `json:"parent" gorm:"index;default:0"`
	CreatedAt time.Time   `json:"created_at" gorm:"index"`
	Langs     []*CateLang `json:"langs" gorm:"foreignKey:CateID"`
	Childs    []*Cate     `json:"-" gorm:"-"`
}
type CateLang struct {
	CateID   int64  `json:"cate_id" gorm:"primaryKey;not null"`
	Lang     string `json:"lang" gorm:"primaryKey;not null"`
	Title    string `json:"title" gorm:"index;not null"`
	Desc     string `json:"desc" gorm:""`
	SeoTitle string `json:"seo_title"`
	SeoDesc  string `json:"seo_desc"`
}

func AddCate(cts []string) {
	// var cccs []*Cate
	for _, cate := range cts {
		cate = strings.ReplaceAll(cate, "，", ",")
		cate = strings.ReplaceAll(cate, "》", ">")
		// cates := strings.Split(cate, ",")
		// for _, v := range cates {
		// 	c := &Cate{

		// 	}
		// 	cccs = append(cccs, c)
		// }
	}
}
