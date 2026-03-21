package audios

import (
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/audios"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetTags(c *gin.Context) {
	s := new(db.ListFinder)
	if err := c.ShouldBindJSON(s); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	tagTotal, tags := audios.GetAudioTags(s)
	response.Success(c, gin.H{
		"total": tagTotal,
		"lists": tags,
	}, "")
}

func RemoveTag(c *gin.Context) {
	type aa struct {
		TagID   int64 `json:"tag_id"`
		AudioID int64 `json:"audio_id"`
	}

	var s aa
	if err := c.ShouldBindJSON(&s); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if s.AudioID < 1 || s.TagID < 1 {
		response.Error(c, http.StatusBadRequest, "数据错误", nil)
		return
	}

	audio := audios.Audio{ID: s.AudioID}
	tag := audios.AudioTag{ID: s.TagID}

	// 删除关联（只删中间表）
	if err := audios.Dbs.Write(func(tx *gorm.DB) error {
		return tx.Model(&audio).Association("Tags").Delete(&tag)
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, nil, "")
}

func EditerTag(c *gin.Context) {
	type et struct {
		AudioID int64    `json:"audio_id"`
		Tags    []string `json:"tags"`
	}

	ets := new(et)
	if err := c.ShouldBindJSON(ets); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if ets.AudioID < 1 {
		response.Error(c, http.StatusBadRequest, "参数错误", nil)
		return
	}

	var audio audios.Audio
	audios.Dbs.DB().First(&audio, ets.AudioID)
	if audio.ID < 1 {
		response.Error(c, http.StatusBadRequest, "音频不存在", nil)
		return
	}

	tags, err := audios.MakerTags(ets.Tags)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := audios.Dbs.Write(func(tx *gorm.DB) error {
		return tx.Model(&audio).Association("Tags").Replace(tags)
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, nil, "")
}

// 批量添加标签
func BatchAdd(c *gin.Context) {
	type tmp struct {
		AIDs []int64  `json:"aids"`
		Tags []string `json:"tags"`
	}

	var req tmp
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if len(req.AIDs) < 1 || len(req.Tags) < 1 {
		response.Error(c, http.StatusBadRequest, "未找到添加的内容", nil)
		return
	}

	var ads []*audios.Audio
	audios.Dbs.DB().Model(&audios.Audio{}).Where("id in ?", req.AIDs).Find(&ads)

	if len(ads) < 1 {
		response.Error(c, http.StatusBadRequest, "未找到添加的内容", nil)
		return
	}

	tags, err := audios.MakerTags(req.Tags)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := audios.Dbs.Write(func(tx *gorm.DB) error {
		for _, audio := range ads {
			if err := tx.Model(audio).Association("Tags").Append(tags); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, nil, "")
}
