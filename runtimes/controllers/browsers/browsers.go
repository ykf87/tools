package browsers

import (
	"fmt"
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/clients"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type TagList struct {
	Q     string `json:"q" form:"q"`
	Limit int    `json:"limit" form:"limit"`
	Page  int    `json:"page" form:"page"`
}

func BrowserTags(c *gin.Context) {
	dt := new(TagList)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	model := db.DB.Model(&clients.BrowserTag{})
	if dt.Q != "" {
		qs := fmt.Sprintf("%%%s%%", dt.Q)
		model = model.Where("name LIKE ?", qs)
	}
	var total int64
	model.Count(&total)
	model = model.Order("id DESC")

	if dt.Limit != -1 {
		if dt.Page < 1 {
			dt.Page = 1
		}
		if dt.Limit < 1 {
			dt.Limit = 20
		}
		model = model.Offset((dt.Page - 1) * dt.Limit).Limit(dt.Limit)
	}

	var list []*clients.BrowserTag
	model.Find(&list)

	rrs, _ := parses.Marshal(list, c)
	response.Success(c, gin.H{"list": rrs, "total": total}, "")
}

type ListStruct struct {
	Page    int     `json:"page" form:"page"`
	Limit   int     `json:"limit" form:"limit"`
	Q       string  `json:"q" form:"q"`
	SortCol string  `json:"scol" form:"scol"`
	By      string  `json:"by" form:"by"`
	Col     string  `json:"col" form:"col"`
	Cval    string  `json:"cval" form:"cval"`
	Tags    []int64 `json:"tags" form:"tags"`
}

func List(c *gin.Context) {
	var l ListStruct
	if err := c.ShouldBindQuery(&l); err != nil {
		response.Error(c, http.StatusNotFound, i18n.T("Error"), nil)
		return
	}

	model := db.DB.Model(&clients.Browser{})
	if l.Q != "" {
		qs := fmt.Sprintf("%%%s%%", l.Q)
		model = model.Where("name LIKE ? OR lang local = ? or lang = ?", qs, qs, l.Q)
	}

	if l.Col != "" && l.Cval != "" {
		model = model.Where(fmt.Sprintf("%s = ?", l.Col), l.Cval)
	}

	if len(l.Tags) > 0 {
		var tagIds []int64
		for _, v := range l.Tags {
			tagIds = append(tagIds, v)
		}
		if len(tagIds) > 0 {
			model = model.Joins("right join browser_to_tags on browser_id = id").Where("browser_to_tags.tag_id in ?", tagIds)
		}
	}

	var total int64
	model.Count(&total)

	sortCol := "id"
	sortBy := "DESC"

	if l.SortCol != "" {
		sortCol = l.SortCol
	}
	if l.By != "" {
		if l.By == "asc" {
			sortBy = "ASC"
		} else {
			sortBy = "DESC"
		}
	}
	if l.Page < 1 {
		l.Page = 1
	}
	if l.Limit < 1 {
		l.Limit = 10
	}

	var ps []*clients.Browser
	model.Order(fmt.Sprintf("%s %s", sortCol, sortBy)).Offset((l.Page - 1) * l.Limit).Limit(l.Limit).Debug().Find(&ps)

	// 处理代理标签
	if len(ps) > 0 {
		clients.SetBrowserTags(ps)
	}

	rs := gin.H{"list": ps, "total": total}
	rsp, _ := parses.Marshal(rs, c)
	response.Success(c, rsp, "")
}

type BrowserAddData struct {
	Id     int64  `json:"id" form:"id"`         // 编辑的时候才有
	Width  int    `json:"width" form:"width"`   // 屏幕宽度
	Height int    `json:"height" form:"height"` // 屏幕高度
	Name   string `json:"name" form:"name"`     // 名称
	Proxy  int64  `json:"proxy" form:"proxy"`   // 代理的id
}
