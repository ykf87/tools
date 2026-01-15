package browserdb

import (
	"fmt"
	"tools/runtimes/db"
)

// 获取所有的可用浏览器
func GetAllBrowsers(page, limit int, query string) []*Browser {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	model := db.DB.Model(&Browser{})
	if query != "" {
		qs := fmt.Sprintf("%%%s%%", query)
		model.Where("name like ?", qs)
	}

	var ls []*Browser
	model.Offset((page - 1) * limit).Limit(limit).Order("id desc").Find(&ls)
	return ls
}

// 通过id列表获取id
func GetBrowsersByIds(ids []int64) []*Browser {
	var bs []*Browser
	db.DB.Model(&Browser{}).Where("id in ?", ids).Find(&bs)
	return bs
}
