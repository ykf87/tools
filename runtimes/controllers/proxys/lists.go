package proxys

import (
	"fmt"
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/proxys"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type ListStruct struct {
	Page    int    `json:"page" form:"page"`
	Limit   int    `json:"limit" form:"limit"`
	Q       string `json:"q" form:"q"`
	SortCol string `json:"scol" form:"scol"`
	By      string `json:"by" form:"by"`
	Col     string `json:"col" form:"col"`
	Cval    string `json:"cval" form:"cval"`
}

// 获取列表
func GetList(c *gin.Context) {
	var l ListStruct
	if err := c.ShouldBindQuery(&l); err != nil {
		response.Error(c, http.StatusNotFound, i18n.T("Error"), nil)
		return
	}

	fmt.Println(l)

	model := db.DB.Model(&proxys.Proxy{})
	if l.Q != "" {
		qs := fmt.Sprintf("%%%s%%", l.Q)
		model = model.Where("name LIKE ? OR remark LIKE ? or local LIKE ?", qs, qs, qs)
	}

	if l.Col != "" && l.Cval != "" {
		model = model.Where(fmt.Sprintf("%s = ?", l.Col), l.Cval)
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

	var ps []*proxys.Proxy
	model.Order(fmt.Sprintf("%s %s", sortCol, sortBy)).Offset((l.Page - 1) * l.Limit).Limit(l.Limit).Debug().Find(&ps)
	rs := gin.H{"list": ps, "total": total}
	rsp, _ := parses.Marshal(rs, c)
	response.Success(c, rsp, "")
}
