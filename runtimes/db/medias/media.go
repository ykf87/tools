package medias

import (
	"os"
	"path/filepath"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/eventbus"

	"gorm.io/gorm"
)

var MediaPath string

type Media struct {
	Id       int64  `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Path     string `json:"path" gorm:"index;default:null" form:"path"`
	Name     string `json:"name" gorm:"default:null" form:"name"`
	Platform string `json:"platform" gorm:"index;default:null" form:"platform"`
	Uuid     string `json:"uuid" gorm:"index;default:null" form:"uuid"`
	Url      string `json:"url" gorm:"default:null" form:"url"`
	Mime     string `json:"mime" gorm:"index" form:"mime"`
	Size     int64  `json:"size" gorm:"index;default:0" form:"size"`
	Addtime  time.Time
}

func init() {
	db.MEDIADB.AutoMigrate(&Media{})

	MediaPath = filepath.Join(config.RuningRoot, "media")
	if _, err := os.Stat(MediaPath); err != nil {
		if err := os.MkdirAll(MediaPath, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func (this *Media) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	if this.Id > 0 {
		err := tx.Model(&Media{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"name":     this.Name,
				"path":     this.Path,
				"platform": this.Platform,
				"url":      this.Url,
				"mime":     this.Mime,
				"size":     this.Size,
			}).Error
		if err != nil {
			return err
		}
		eventbus.Bus.Publish("media_save", this)
		return nil
	} else {
		this.Addtime = time.Now()
		return tx.Create(this).Error
	}
}
