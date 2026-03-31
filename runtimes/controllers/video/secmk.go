package video

import (
	"tools/runtimes/db"
	"tools/runtimes/db/audios"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

// 二创的数据
func SecmkData(c *gin.Context) {
	// uts := medias.GetTags() // 用户标签
	//

	_, audioSlice := audios.GetList(&db.ListFinder{
		Limit: 12,
		Page:  1,
	})
	_, audioTagSlice := audios.GetAudioTags(&db.ListFinder{
		Page:  1,
		Limit: 500,
	})

	response.Success(c, gin.H{
		"audios":     audioSlice,
		"audio_tags": audioTagSlice,
		"user_tags":  medias.GetTags(),
	}, "")
}
