package audios

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/db/audios"
	"tools/runtimes/ffmpeg"
	"tools/runtimes/response"
	"tools/runtimes/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Cut(c *gin.Context) {
	type cutobj struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Title string  `json:"title"`
	}
	type reqobj struct {
		ID   int64    `json:"id"`
		Src  string   `json:"src"`
		Cuts []cutobj `json:"cuts"`
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
	if !strings.Contains(req.Src, ".") {
		response.Error(c, http.StatusBadRequest, "音频文件路径不正确", nil)
		return
	}
	if len(req.Cuts) < 1 {
		response.Error(c, http.StatusBadRequest, "未找到需裁切区域", nil)
		return
	}

	row := audios.GetRow(req.ID)
	if row == nil || row.ID < 1 {
		response.Error(c, http.StatusBadRequest, "音频不存在", nil)
		return
	}

	req.Src = storage.Load(row.StorageType).URL(row.Name)

	tempPath := config.FullPath(config.MEDIAROOT, ".tmp")
	filename := filepath.Base(req.Src)
	ttt := strings.Split(filename, ".")
	ext := ttt[len(ttt)-1]
	justName := strings.TrimRight(filename, "."+ext)

	var creater []*audios.Audio
	for k, v := range req.Cuts {
		tmpname := filepath.Join(tempPath, fmt.Sprintf("%s-%d.%s", justName, (k+1), ext))
		fmt.Println(tmpname, "----")
		_, _, err := ffmpeg.RunFfmpeg(true,
			"-i", req.Src,
			"-ss", fmt.Sprintf("%f", v.Start),
			"-to", fmt.Sprintf("%f", v.End),
			"-acodec", "libmp3lame",
			tmpname,
			"-y",
		)
		fmt.Println("11111")
		if err != nil {
			fmt.Println("生成失败")
			response.Error(c, http.StatusBadRequest, err.Error(), nil)
			return
		}

		// newSrc, err := storage.Load(row.StorageType).PutStr(tmpname)
		// if err != nil {
		// 	fmt.Println("上传失败!")
		// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
		// 	return
		// }
		// fmt.Println("22222222222")
		title := v.Title
		if title == "" {
			title = fmt.Sprintf("%d:%s", k+1, row.Title)
		}
		ad, err := audios.AddAudio(tmpname, title)
		if err != nil {
			response.Error(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		ad.Origin = row.Name
		creater = append(creater, ad)
	}
	if len(creater) > 0 {
		fmt.Println("33333")
		if err := audios.Dbs.Write(func(tx *gorm.DB) error {
			return tx.Create(creater).Error
		}); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
	}
	// _, _, err := ffmpeg.RunFfmpeg(true,
	// 	"-i", req.Src,
	// 	"-ss", fmt.Sprintf("%f", req.Start),
	// 	"-to", fmt.Sprintf("%f", req.End),
	// 	"-c", "copy",
	// 	tempsrc,
	// 	"-y",
	// )
	// if err != nil {
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }

	// newSrc, err := storage.Load(row.StorageType).PutStr(tempsrc)
	// if err != nil {
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }
	// if row.Origin == "" {
	// 	row.Origin = row.Name
	// } else {
	// 	storage.Load(row.StorageType).Delete(row.Name)
	// }
	// row.Name = newSrc

	// if err := audios.Dbs.Write(func(tx *gorm.DB) error {
	// 	return row.Save(row, tx)
	// }); err != nil {
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }
	response.Success(c, nil, "")
}
