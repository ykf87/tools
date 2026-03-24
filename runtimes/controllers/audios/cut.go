package audios

import (
	"fmt"
	"net/http"
	"path/filepath"
	"tools/runtimes/config"
	"tools/runtimes/db/audios"
	"tools/runtimes/ffmpeg"
	"tools/runtimes/response"
	"tools/runtimes/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Cut(c *gin.Context) {
	type reqobj struct {
		ID    int64   `json:"id"`
		Src   string  `json:"src"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	}
	var req reqobj

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if req.ID < 1 || req.Src == "" {
		response.Error(c, http.StatusBadRequest, "数据不完整", nil)
		return
	}
	row := audios.GetRow(req.ID)
	if row == nil || row.ID < 1 {
		response.Error(c, http.StatusBadRequest, "音频不存在", nil)
		return
	}

	tempsrc := config.FullPath(config.MEDIAROOT, ".tmp", filepath.Base(req.Src))

	_, _, err := ffmpeg.RunFfmpeg(true,
		"-i", req.Src,
		"-ss", fmt.Sprintf("%f", req.Start),
		"-to", fmt.Sprintf("%f", req.End),
		"-c", "copy",
		tempsrc,
		"-y",
	)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newSrc, err := storage.Load(row.StorageType).PutStr(tempsrc)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if row.Origin == "" {
		row.Origin = row.Name
	} else {
		storage.Load(row.StorageType).Delete(row.Name)
	}
	row.Name = newSrc

	if err := audios.Dbs.Write(func(tx *gorm.DB) error {
		return row.Save(row, tx)
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, row, "")
}
