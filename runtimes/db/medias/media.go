package medias

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/downloader"
	"tools/runtimes/funcs"
	"tools/runtimes/ratelimit"

	"gorm.io/gorm"
)

type Media struct {
	Id         int64     `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Path       string    `json:"path" gorm:"index;default:null" form:"path"`         // 相对于存储根目录的纯路径
	Name       string    `json:"name" gorm:"default:null;index" form:"name"`         // 真实文件名称,不含路径
	Title      string    `json:"title" gorm:"default:null" form:"title"`             // 显示的标题
	Md5        string    `json:"md5" gorm:"uniqueIndex; not null"`                   // 文件的md5值
	UrlMd5     string    `json:"url_md5" gorm:"uniqueIndex; not null"`               // 下载地址的md5
	Platform   string    `json:"platform" gorm:"index;default:null" form:"platform"` // 平台
	UserId     int64     `json:"user_id" gorm:"index;default:0"`                     // MediaUser id
	VideoID    string    `json:"video_id" gorm:"index"`                              // 自动下载才有的短视频的唯一id
	Url        string    `json:"url" gorm:"default:null" form:"url"`                 // 下载地址
	Mime       string    `json:"mime" gorm:"index" form:"mime"`                      // mime
	Size       int64     `json:"size" gorm:"index;default:0" form:"size"`            // 大小
	Removed    int       `json:"removed" gorm:"default:0;index"`                     // 是否删除
	TempFile   string    `json:"temp_file"`                                          // 删除后的文件路径
	RemoveTime int64     `json:"remove_time" gorm:"default:0;index"`                 // 删除时间
	Filetime   int64     `json:"filetime" gorm:"index;default:0" form:"filetime"`    // 文件最后修改日期
	Addtime    time.Time // 本数据添加日期
	db.BaseModel
}

var dbs *db.SQLiteWriter
var getInfoLimit *ratelimit.Limiter

func init() {
	dbs = db.MEDIADB
	dbs.DB().AutoMigrate(&Media{})
	dbs.DB().AutoMigrate(&MediaUser{})
	dbs.DB().AutoMigrate(&MediaUserTag{})
	dbs.DB().AutoMigrate(&MediaUserToTag{})
	dbs.DB().AutoMigrate(&MediaUserToClient{})
	dbs.DB().AutoMigrate(&MediaUserProxy{})
	dbs.DB().AutoMigrate(&MediaUserDay{})
	go runstart()

	getInfoLimit = ratelimit.New(
		ratelimit.WithLimit(5, 10*time.Second), // 10秒5次
		ratelimit.WithConcurrency(2),           // 最大2并发
		ratelimit.WithQueue(100),               // 队列长度
	)

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
		mu.Addtime = time.Now().Unix()
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

// 删除媒体文件
func DeleteMediaFiles(ids []int64) error {
	var ms []*Media
	if err := dbs.DB().Model(&Media{}).Where("id in ? and removed = 0", ids).Find(&ms).Error; err != nil {
		return err
	}
	if len(ms) < 1 {
		return errors.New("找不到文件")
	}

	return dbs.Write(func(tx *gorm.DB) error {
		if err := tx.Model(&Media{}).Where("id in ?", ids).UpdateColumns(map[string]any{
			"removed":     1,
			"remove_time": time.Now().Unix(),
			"temp_file":   gorm.Expr("? || '/' || name", config.TRASHED),
		}).Error; err != nil {
			return err
		}

		for _, m := range ms {
			origin := config.FullPath(filepath.Join(config.MEDIAROOT, m.Path, m.Name))
			if _, err := os.Stat(origin); err != nil {
				continue
			}
			if err := os.Rename(
				origin,
				config.FullPath(config.TRASHED, m.Name),
			); err != nil {
				return err
			}
		}

		return nil
	})
	// origin := config.FullPath(filepath.Join(config.MEDIAROOT, m.Path, m.Name))
	// to := config.FullPath(config.TRASHED, m.Name)
	// return dbs.Write(func(tx *gorm.DB) error {
	// 	m.Removed = 1
	// 	m.RemoveTime = time.Now().Unix()
	// 	m.TempFile = filepath.Join(config.TRASHED, m.Name)
	// 	if err := m.Save(m, tx); err != nil {
	// 		return err
	// 	}
	// 	if err := os.Rename(origin, to); err != nil {
	// 		return err
	// 	}
	// 	return nil
	// })
}
