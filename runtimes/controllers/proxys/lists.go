package proxys

import (
	"fmt"
	"net/http"
	"strings"
	"tools/runtimes/db"
	"tools/runtimes/db/proxys"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type ListStruct struct {
	Page    int     `json:"page" form:"page"`
	Limit   int     `json:"limit" form:"limit"`
	Q       string  `json:"q" form:"q"`
	SortCol string  `json:"scol" form:"scol"`
	By      string  `json:"by" form:"by"`
	Col     string  `json:"col" form:"col"`
	Cval    string  `json:"cval" form:"cval"`
	Sub     int64   `json:"sub" form:"sub"`
	Tags    []int64 `json:"tags" form:"tags"`
}

// 获取列表
func GetList(c *gin.Context) {
	var l ListStruct
	if err := c.ShouldBindQuery(&l); err != nil {
		response.Error(c, http.StatusNotFound, i18n.T("Error"), nil)
		return
	}

	model := db.DB.Model(&proxys.Proxy{})
	if l.Q != "" {
		qs := fmt.Sprintf("%%%s%%", l.Q)
		model = model.Where("name LIKE ? OR remark LIKE ? or local = ?", qs, qs, l.Q)
	}

	if l.Col != "" && l.Cval != "" {
		model = model.Where(fmt.Sprintf("%s = ?", l.Col), l.Cval)
	}

	if l.Sub < 1 {
		model = model.Where("subscribe = 0")
	} else {
		model = model.Where("subscribe = ?", l.Sub)
	}

	if len(l.Tags) > 0 {
		var tagIds []int64
		for _, v := range l.Tags {
			tagIds = append(tagIds, v)
		}
		if len(tagIds) > 0 {
			model = model.Joins("right join proxy_tags on proxy_id = id").Where("proxy_tags.tag_id in ?", tagIds)
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

	if l.Limit != -1 {
		if l.Page < 1 {
			l.Page = 1
		}
		if l.Limit < 1 {
			l.Limit = 10
		}
		model = model.Offset((l.Page - 1) * l.Limit).Limit(l.Limit)
	}

	var ps []*proxys.Proxy
	model.Order(fmt.Sprintf("%s %s", sortCol, sortBy)).Find(&ps)

	// 处理代理标签
	if len(ps) > 0 {
		proxys.SetTags(ps)
	}
	for _, v := range ps {
		if cc := v.IsStart(); cc != nil {
			v.IsRuning = 1
			v.ListerAddr = cc.Listened()
		}
	}

	rs := gin.H{"list": ps, "total": total}
	rsp, _ := parses.Marshal(rs, c)
	response.Success(c, rsp, "")
}

// 获取一行数据
func GetRow(c *gin.Context) {
	id := c.Param("id")
	px := proxys.GetById(id)

	if px == nil {
		response.Error(c, http.StatusNotFound, "", nil)
		return
	}
	pxc, _ := parses.Marshal(px, c)
	response.Success(c, pxc, "Success")
}

// 启动
func Start(c *gin.Context) {
	ids := c.Param("id")
	for _, id := range strings.Split(ids, ","){
		pc := proxys.GetById(id)
		if pc != nil && pc.Id > 0 {
			if _, err := pc.Start(false); err != nil {
				response.Error(c, http.StatusOK, err.Error(), nil)
				return
			}
		}
	}

	response.Success(c, nil, "Success")
}

// 停止
func Stop(c *gin.Context) {
	ids := c.Param("id")
	for _, id := range strings.Split(ids, ","){
		pc := proxys.GetById(id)
		if pc != nil && pc.Id > 0 {
			if err := pc.Stop(true); err != nil {
				response.Error(c, http.StatusOK, err.Error(), nil)
				return
			}
		}
	}
	response.Success(c, nil, "Success")
}
