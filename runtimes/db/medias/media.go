package medias

import (
	"time"
	"tools/runtimes/db"
	"tools/runtimes/eventbus"

	"gorm.io/gorm"
)

type MediaUser struct {
	Id       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string `json:"name" gorm:"index;default:null"`
	Cover    string `json:"cover" gorm:"default:null"`
	Platform string `json:"platform" gorm:"index:plu;not null"`
	Uuid     string `json:"uuid" gorm:"index:plu;not null"`
}

type Media struct {
	Id       int64     `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Path     string    `json:"path" gorm:"index;default:null" form:"path"` // 相对于存储根目录的纯路径
	Name     string    `json:"name" gorm:"default:null" form:"name"`       // 真实文件名称,不含路径
	Title    string    `json:"title" gorm:"default:null" form:"title"`     // 显示的标题
	Md5      string    `json:"md5" gorm:"uniqueIndex"`
	Platform string    `json:"platform" gorm:"index;default:null" form:"platform"`
	UserId   int64     `json:"user_id" gorm:"index;default:0"`     // MediaUser id
	Url      string    `json:"url" gorm:"default:null" form:"url"` // 下载地址
	Mime     string    `json:"mime" gorm:"index" form:"mime"`
	Size     int64     `json:"size" gorm:"index;default:0" form:"size"`         //大小
	Filetime int64     `json:"filetime" gorm:"index;default:0" form:"filetime"` // 文件最后修改日期
	Addtime  time.Time // 本数据添加日期
}

func init() {
	db.MEDIADB.AutoMigrate(&Media{})
	db.MEDIADB.AutoMigrate(&MediaUser{})
}

func MkerMediaUser(platform, uid, cover, name string) *MediaUser {
	var mu *MediaUser
	if err := db.MEDIADB.Model(&MediaUser{}).Where("platform = ? and uuid = ?", platform, uid).First(mu).Error; err != nil {
		mu.Cover = cover
		mu.Name = name
		mu.Platform = platform
		mu.Uuid = uid
		mu.Save(nil)
	}
	return mu
}

func (this *MediaUser) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.MEDIADB
	}

	if this.Id > 0 {
		err := tx.Model(&Media{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"platform": this.Platform,
				"name":     this.Name,
				"cover":    this.Cover,
				"uuid":     this.Uuid,
			}).Error
		if err != nil {
			return err
		}
		eventbus.Bus.Publish("media_save", this)
		return nil
	} else {
		return tx.Create(this).Error
	}
}

func (this *Media) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.MEDIADB
	}

	if this.Id > 0 {
		err := tx.Model(&Media{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"title":    this.Title,
				"name":     this.Name,
				"path":     this.Path,
				"md5":      this.Md5,
				"platform": this.Platform,
				"url":      this.Url,
				"mime":     this.Mime,
				"size":     this.Size,
				"filetime": this.Filetime,
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
