package medias

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/downloader"
	"tools/runtimes/eventbus"
	"tools/runtimes/funcs"

	"gorm.io/gorm"
)

type Media struct {
	Id       int64     `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Path     string    `json:"path" gorm:"index;default:null" form:"path"`         // 相对于存储根目录的纯路径
	Name     string    `json:"name" gorm:"default:null;index" form:"name"`         // 真实文件名称,不含路径
	Title    string    `json:"title" gorm:"default:null" form:"title"`             // 显示的标题
	Md5      string    `json:"md5" gorm:"uniqueIndex; not null"`                   // 文件的md5值
	UrlMd5   string    `json:"url_md5" gorm:"uniqueIndex; not null"`               // 下载地址的md5
	Platform string    `json:"platform" gorm:"index;default:null" form:"platform"` // 平台
	UserId   int64     `json:"user_id" gorm:"index;default:0"`                     // MediaUser id
	VideoID  string    `json:"video_id" gorm:"index"`                              // 自动下载才有的短视频的唯一id
	Url      string    `json:"url" gorm:"default:null" form:"url"`                 // 下载地址
	Mime     string    `json:"mime" gorm:"index" form:"mime"`                      // mime
	Size     int64     `json:"size" gorm:"index;default:0" form:"size"`            //大小
	Filetime int64     `json:"filetime" gorm:"index;default:0" form:"filetime"`    // 文件最后修改日期
	Addtime  time.Time // 本数据添加日期
}

var dbs *db.SQLiteWriter

func init() {
	dbs = db.MEDIADB
	dbs.DB().AutoMigrate(&Media{})
	dbs.DB().AutoMigrate(&MediaUser{})
	dbs.DB().AutoMigrate(&MediaUserTag{})
	dbs.DB().AutoMigrate(&MediaUserToTag{})
	dbs.DB().AutoMigrate(&MediaUserToClient{})
	dbs.DB().AutoMigrate(&MediaUserProxy{})
	dbs.DB().AutoMigrate(&MediaUserDay{})
	runstart()

	// var mus []*MediaUser
	// dbs.Model(&MediaUser{}).Where("autoinfo = 1 or auto_download = 1").Find(&mus)
	// for _, v := range mus {
	// 	autoLoaderUser.Store(v.Id, v)
	// }

	// go autoStart()
}

func GetDb() *db.SQLiteWriter {
	return dbs
}

func MkerMediaUser(platform, uid, cover, name, proxy, searchID string, adminID int64) *MediaUser {
	mu := new(MediaUser)
	// fmt.Println(searchID, "mkuser------------------")
	if err := dbs.DB().Model(&MediaUser{}).Where("platform = ? and uuid = ?", platform, uid).First(mu).Error; err != nil {
		exts := "png"
		dl := downloader.NewDownloader(proxy, nil, nil)
		if ext, err := dl.GetUrlFileExt(cover); err == nil {
			exts = ext
		}

		dest := fmt.Sprintf("avatar/%s.%s", funcs.Md5String(cover), exts)
		fullPath := filepath.Join(config.MEDIAROOT, dest)
		dirRoot := filepath.Dir(fullPath)
		if _, err := os.Stat(dirRoot); err != nil {
			os.MkdirAll(dirRoot, os.ModePerm)
		}
		mu.Cover = cover

		if err := dl.Download(cover, fullPath); err == nil {
			mu.Cover = dest
		}
		mu.Name = name
		mu.Platform = platform
		mu.Uuid = uid
		mu.AdminID = adminID
		mu.Account = searchID
	} else {
		mu.Account = searchID
	}

	dbs.Write(func(tx *gorm.DB) error {
		return mu.Save(mu, tx)
	})
	return mu
}

func (this *Media) Save(tx *db.SQLiteWriter) error {
	if tx == nil {
		tx = dbs
	}

	if this.Id > 0 {
		err := tx.Write(func(txx *gorm.DB) error {
			return txx.Model(&Media{}).Where("id = ?", this.Id).
				Updates(map[string]any{
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
		})
		if err != nil {
			return err
		}
		eventbus.Bus.Publish("media_save", this)
		return nil
	} else {
		this.Addtime = time.Now()
		return tx.Write(func(txx *gorm.DB) error {
			return txx.Create(this).Error
		})
	}
}

func GetMediasUserFromName(names []string) map[string]*MediaUser {
	var mmus []*Media

	resp := make(map[string]*MediaUser)
	if err := dbs.DB().Model(&Media{}).Where("name in ?", names).Find(&mmus).Error; err == nil {
		var ids []int64
		// idNames := make(map[int64]string)
		for _, v := range mmus {
			if v.UserId > 0 {
				ids = append(ids, v.UserId)
				// idNames[v.UserId] = v.Name
			}
		}

		var mmuus []*MediaUser
		if len(ids) > 0 {
			dbs.DB().Model(&MediaUser{}).Where("id in ?", ids).Find(&mmuus)
			sdsd := make(map[int64]*MediaUser)
			for _, v := range mmuus {
				sdsd[v.Id] = v
			}

			for _, v := range mmus {
				if vvu, ok := sdsd[v.UserId]; ok {
					resp[v.Name] = vvu
				}
			}
		}
	}
	return resp
}

// 通过url的md5获取行
func GerUrlMd5Row(md5 string) *Media {
	md := new(Media)
	dbs.DB().Model(&Media{}).Where("url_md5 = ?", md5).First(md)
	return md
}
