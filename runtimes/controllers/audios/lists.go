package audios

import (
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/audios"
	"tools/runtimes/response"
	"tools/runtimes/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func List(c *gin.Context) {
	var geter db.ListFinder

	if err := c.ShouldBindJSON(&geter); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	total, lists := audios.GetList(&geter)
	tagTotal, tags := audios.GetAudioTags(&db.ListFinder{
		Page: 1,
	})

	response.Success(c, gin.H{
		"total":     total,
		"lists":     lists,
		"tag_total": tagTotal,
		"tags":      tags,
	}, "")
}

func Delete(c *gin.Context) {
	type idss struct {
		IDs []int64 `json:"ids"`
	}

	var ids idss
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if len(ids.IDs) < 1 {
		response.Success(c, nil, "内容为空")
		return
	}

	var has []*audios.Audio
	audios.Dbs.DB().Where("id in ?", ids.IDs).Find(&has)

	if err := audios.Dbs.Write(func(tx *gorm.DB) error {
		if err := tx.Table("audio_tag_relations").Where("audio_id in ?", ids.IDs).Delete(nil).Error; err != nil {
			return err
		}
		return tx.Where("id in ?", ids.IDs).Delete(&audios.Audio{}).Error
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	for _, v := range has {
		storage.Load(v.StorageType).Delete(v.Name)
	}
	response.Success(c, nil, "")
}
