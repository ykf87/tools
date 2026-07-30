package spider

import (
	"errors"
	"io"
	"path/filepath"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/funcs"
	"tools/runtimes/storage"
)

type SpiderJs struct { // 如果有修改字段,需要更新Save方法
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name      string    `json:"name" gorm:"index;not null" form:"id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	Content   string    `json:"content"`
	db.BaseModel
}

type SpiderMedia struct {
	ID       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Md5      string `json:"md5" gorm:"index"`
	Url      string `json:"url" gorm:"not null"`
	Uri      string `json:"uri" gorm:"not null"`
	FileName string `json:"file_name"`
}

func init() {
	db.DB.DB().AutoMigrate(&SpiderJs{}, &SpiderMedia{})
}

func GetSpiderJses() []*SpiderJs {
	var jses []*SpiderJs
	db.DB.DB().Order("created_at desc").Find(&jses)
	return jses
}

func GetSpiderMedia(url string) ([]byte, string, error) {
	md5 := funcs.Md5String(url)
	var row SpiderMedia
	db.DB.DB().Where("md5 = ?", md5).First(&row)

	if row.ID > 0 {
		if o, err := storage.Load("minio").Get(row.Uri); err == nil {
			if dt, err := io.ReadAll(o); err == nil {
				return dt, row.FileName, nil
			}
		}
	}
	return nil, "", errors.New("not found")
}

func SaveSpiderMedia(url, file string) error {
	uri, err := storage.Load("minio").PutStr(file, false)
	if err == nil {
		row := &SpiderMedia{
			Md5:      funcs.Md5String(url),
			Url:      url,
			Uri:      uri,
			FileName: filepath.Base(file),
		}
		if err := db.DB.DB().Model(&SpiderMedia{}).Create(row).Error; err != nil {
			return err
		}
		return nil
	}
	return err
}
