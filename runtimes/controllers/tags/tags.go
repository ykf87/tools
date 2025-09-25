package tags

import (
	"fmt"
	"net/http"
	"strconv"
	"tools/runtimes/db"
	"tools/runtimes/db/tag"
	"tools/runtimes/i18n"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type Data struct {
	Q     string `json:"q" form:"q"`
	Limit int    `json:"limit" form:"limit"`
	Page  int    `json:"page" form:"page"`
}

// 获取标签列表
func List(c *gin.Context) {
	dt := new(Data)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	model := db.DB.Model(&tag.Tag{})
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

	var list []*tag.Tag
	model.Find(&list)

	rrs, _ := parses.Marshal(list, c)
	response.Success(c, gin.H{"list": rrs, "total": total}, "")
}

// 添加标签
func Add(c *gin.Context) {
	dt := new(tag.Tag)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	if dt.Id > 0 {
		response.Error(c, http.StatusBadGateway, i18n.T("Modification is not supported, please delete and add again"), nil)
		return
	}

	if err := db.DB.Create(dt).Error; err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	response.Success(c, dt, "")
}

// 删除标签
func Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id > 0 {
		db.DB.Where("id = ?", id).Delete(&tag.Tag{})
	}

	response.Success(c, nil, "")
}
