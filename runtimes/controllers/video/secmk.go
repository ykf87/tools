package video

import (
	"fmt"
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/audios"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

// 二创的数据
func SecmkData(c *gin.Context) {
	// _, audioSlice := audios.GetList(&db.ListFinder{
	// 	Limit: 12,
	// 	Page:  1,
	// })
	_, audioTagSlice := audios.GetAudioTags(&db.ListFinder{
		Page:  1,
		Limit: 500,
	})

	response.Success(c, gin.H{
		// "audios":     audioSlice,
		"audio_tags": audioTagSlice,
		"user_tags":  medias.GetTags(),
	}, "")
}

// 按条件查询视频
type smData struct {
	UserIDs          []int64 `json:"user_ids"`
	UserTags         []int64 `json:"user_tags"`
	SearchMediaTitle string  `json:"search_media_title"`
	Page             int     `json:"page"`
	Limit            int     `json:"limit"`
}

func SearchMedias(c *gin.Context) {
	var req smData
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if len(req.UserTags) > 0 {
		uids := medias.GetMUIDsByTagIDs(req.UserTags)
		req.UserIDs = append(req.UserIDs, uids...)
	}

	model := medias.GetDb().DB().Model(&medias.Media{}).
		Where("removed = 0").Where(`
    EXISTS (
        SELECT 1 FROM media_files
        WHERE media_files.m_id = media.id
        AND mime_type LIKE ?
    )
`, "video/%").Preload("Files", "mime_type LIKE ?", "video/%")

	if len(req.UserIDs) > 0 {
		model = model.Where("user_id in ?", req.UserIDs)
	}
	if req.SearchMediaTitle != "" {
		model = model.Where("title like ?", fmt.Sprintf("%%%s%%", req.SearchMediaTitle))
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	}

	var mds []*medias.Media
	var total int64
	model.Count(&total)
	model.Offset((req.Page - 1) * req.Limit).Limit(req.Limit).Order("id DESC").Find(&mds)
	response.Success(c, gin.H{
		"total": total,
		"lists": mds,
	}, "")
}

// 视频制作
func MakerVideos(c *gin.Context) {

}
