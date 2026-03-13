package mediasec

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"tools/runtimes/config"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"
	"tools/runtimes/storage"

	"github.com/gin-gonic/gin"
)

func GetVideos(c *gin.Context) {
	type abc struct {
		UserID []int64 `json:"user_id" form:"user_id"` // 用户id
		GetNum int     `json:"get_num" form:"get_num"` // 获取数量
	}

	var getObj abc
	if err := c.ShouldBindJSON(&getObj); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if len(getObj.UserID) < 1 {
		response.Error(c, http.StatusBadRequest, "请传入用户id", nil)
		return
	}

	if getObj.GetNum < 1 {
		getObj.GetNum = 1
	}

	var mds []*medias.Media
	if err := medias.GetDb().DB().
		Model(&medias.Media{}).Preload("Files", "mime_type like ?", "%%%s%", "video").
		Where("user_id in ? and removed = 0 and used = 0", getObj.UserID).
		Order("id ASC").Limit(getObj.GetNum).
		Find(&mds).Error; err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if len(mds) < 1 {
		response.Error(c, http.StatusBadRequest, "找不到新的视频", nil)
		return
	}

	for _, v := range mds {
		fl := len(v.Files)
		if fl != 1 {
			continue
		}

		rd, err := storage.Load("").Get(v.Files[0].FileName)
		if err != nil {
			continue
		}
		saveto := config.FullPath(config.MEDIAROOT, ".tmp")
		if _, err := os.Stat(saveto); err != nil {
			if err := os.MkdirAll(saveto, os.ModePerm); err != nil {
				response.Error(c, http.StatusBadRequest, err.Error(), nil)
				return
			}
		}
		finalFilaPath := fmt.Sprintf("%s/%s", saveto, filepath.Base(v.Files[0].FileName))
		file, err := os.Create(finalFilaPath)
		if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error(), nil)
			return
		}

		if _, err = io.Copy(file, rd); err != nil {
			file.Close()
			response.Error(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		file.Close()

		// storage.Load("").Get(v.Files[0].FileName)
		// wc := wmimage.DetectFromFile(,100)
	}

}
