package product

import (
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/langs"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetLangs(c *gin.Context) {
	var req db.ListFinder
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	list, total := langs.GetLangs(req)
	response.Success(c, gin.H{
		"lists": list,
		"total": total,
	}, "")
}

func AddLang(c *gin.Context) {
	var addlangs []langs.Lang

	if err := c.ShouldBindJSON(&addlangs); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := langs.DB.Write(func(tx *gorm.DB) error {
		return tx.Create(addlangs).Error
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, nil, "")
}

func DeleteLangs(c *gin.Context) {
	type idss struct {
		IDs []int64 `json:"ids" form:"ids"`
	}
	var ids idss

	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := langs.DB.Write(func(tx *gorm.DB) error {
		return tx.Where("id in ?", ids.IDs).Delete(&langs.Lang{}).Error
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, nil, "")
}
