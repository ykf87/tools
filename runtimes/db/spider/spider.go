package spider

import (
	"bytes"
	"errors"
	"path/filepath"
	"tools/runtimes/db"
	"tools/runtimes/funcs"
	"tools/runtimes/storage"
)

type SpiderMedia struct {
	ID       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Md5      string `json:"md5" gorm:"index"`
	Url      string `json:"url" gorm:"not null"`
	Uri      string `json:"uri" gorm:"not null"`
	FileName string `json:"file_name"`
}

var DB = db.PRODUCTDB

func init() {
	DB.DB().AutoMigrate(&SpiderJs{}, &SpiderMedia{})
}

func GetSpiderJses() []*SpiderJs {
	var jses []*SpiderJs
	DB.DB().Order("created_at desc").Find(&jses)
	return jses
}

func GetSpiderMedia(url string) (string, int64, error) {
	md5 := funcs.Md5String(url)
	var row SpiderMedia
	DB.DB().Where("md5 = ?", md5).First(&row)

	if row.ID > 0 {
		return row.Uri, row.ID, nil
	}
	return "", 0, errors.New("not found")
}

func SaveSpiderMedia(url, fileName string, dt *bytes.Reader) (int64, string, error) {
	// uri, err := storage.Load("minio").PutStr(file, false)
	urlmd5 := funcs.Md5String(url)
	var dbrow SpiderMedia
	DB.DB().Where("md5 = ?", urlmd5).First(&dbrow)
	if dbrow.ID > 0 {
		return dbrow.ID, dbrow.Uri, nil
	}

	uri, err := storage.Load("minio").Put(dt, nil)
	if err == nil {
		row := SpiderMedia{
			Md5:      funcs.Md5String(url),
			Url:      url,
			Uri:      uri,
			FileName: filepath.Base(fileName),
		}
		if err := DB.DB().Model(&SpiderMedia{}).Create(&row).Error; err != nil {
			return 0, "", err
		}
		return row.ID, uri, nil
	}
	return 0, "", err
}
