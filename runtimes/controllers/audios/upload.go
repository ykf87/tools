package audios

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/db/audios"
	"tools/runtimes/db/medias"
	"tools/runtimes/response"
	"tools/runtimes/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Uploads(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, 400, err.Error(), nil)
		return
	}
	defer file.Close()
	storageType := c.PostForm("storage")
	if storageType == "" {
		storageType = config.DefStorage
	}

	contentType := header.Header.Get("Content-Type")
	if strings.Contains(contentType, "audio") == false && strings.Contains(contentType, "video") == false {
		response.Error(c, 400, "请上传音频或者视频文件", nil)
		return
	}

	tmpFile := config.FullPath(config.MEDIAROOT, "tmp", header.Filename)
	bsdir := filepath.Dir(tmpFile)
	if _, err := os.Stat(bsdir); err != nil {
		os.MkdirAll(bsdir, os.ModePerm)
	}
	out, err := os.Create(tmpFile)
	if err != nil {
		response.Error(c, 400, "保存文件失败", nil)
		return
	}
	_, err = io.Copy(out, file)
	if err != nil {
		out.Close()
		response.Error(c, 400, "写入文件失败:"+err.Error(), nil)
		return
	}
	out.Close()

	ad, err := audios.AddAudio(tmpFile, strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)))
	os.Remove(tmpFile)
	if err != nil {
		response.Error(c, 400, "上传失败:"+err.Error(), nil)
		return
	}
	if err := audios.Dbs.Write(func(tx *gorm.DB) error {
		return ad.Save(ad, tx)
	}); err != nil {
		response.Error(c, 400, "保存失败:"+err.Error(), nil)
		return
	}

	response.Success(c, nil, "")
}

func UploadFromMedia(c *gin.Context) {
	md := medias.GetRow(c.Param("id"))
	if len(md.Files) != 1 {
		response.Error(c, http.StatusBadGateway, "未找到文件内容", nil)
		return
	}
	filename := storage.Load(md.Files[0].FileSystem).URL(md.Files[0].FileName)
	// fmt.Println(filename)
	cc, err := audios.AddAudio(filename, md.Title)
	if err != nil {
		response.Error(c, http.StatusBadGateway, "未找到文件内容", nil)
		return
	}
	cc.UserID = md.UserId
	cc.MID = md.Id

	if err := audios.Dbs.Write(func(tx *gorm.DB) error {
		return cc.Save(cc, tx)
	}); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	response.Success(c, nil, "")
}
