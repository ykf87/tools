package medias

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/downloader"
	"tools/runtimes/funcs"
	"tools/runtimes/mainsignal"
	"tools/runtimes/proxy"
	"tools/runtimes/ratelimit"
	"tools/runtimes/storage"
	"tools/runtimes/videos/downloader/parser"

	"gorm.io/gorm"
)

type Media struct {
	Id           int64        `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	PathID       int64        `json:"path_id" gorm:"index;default:0"`                          // 路径的id
	Path         string       `json:"path" gorm:"index;" form:"path"`                          // 概念上的path,只是在显示端使用,并不会有真实的路径创建
	Title        string       `json:"title" gorm:"default:null" form:"title"`                  // 显示的标题
	Platform     string       `json:"platform" gorm:"index;default:null" form:"platform"`      // 平台
	UserId       int64        `json:"user_id" gorm:"index;default:0"`                          // MediaUser id
	VideoID      string       `json:"video_id" gorm:"index"`                                   // 自动下载才有的短视频的唯一id
	Url          string       `json:"url" gorm:"default:null" form:"url"`                      // 下载地址
	UrlMd5       string       `json:"url_md5" gorm:"uniqueIndex; not null"`                    // 下载地址的md5
	AdminID      int64        `json:"admin_id" gorm:"default:0;index"`                         // 管理员id
	Cover        string       `json:"cover" gorm:"default:null"`                               // 封面
	CoverStorage string       `json:"-"`                                                       // 封面适用的存储器
	Sizes        int64        `json:"sizes" gorm:"index;default:0" form:"size"`                // 大小
	Numbers      int          `json:"numbers" gorm:"index;default:0"`                          // 文件数量
	Removed      int          `json:"removed" gorm:"default:0;index"`                          // 是否删除
	Addtime      time.Time    `json:"addtime" gorm:"index"`                                    // 本数据添加日期
	Files        []*MediaFile `json:"files" gorm:"foreignKey:MID;constraint:OnDelete:CASCADE"` // 文件列表
	User         *MediaUser   `json:"user" gorm:"foreignKey:UserId;references:Id"`             // 用户
	Used         int          `json:"used" gorm:"default:0;index"`                             // 是否通过获取接口使用过
	// TempFile   string      `json:"temp_file"`                                          // 删除后的文件路径
	// RemoveTime int64       `json:"remove_time" gorm:"default:0;index"`                 // 删除时间
	// Filetime   int64       `json:"filetime" gorm:"index;default:0" form:"filetime"`    // 文件最后修改日期
	// Mime       string      `json:"mime" gorm:"index" form:"mime"`                      // mime
	// Md5      string `json:"md5" gorm:"uniqueIndex; not null"`                   // 文件的md5值
	// Name       string      `json:"name" gorm:"default:null;index" form:"name"`         // 真实文件名称,不含路径
	// db.BaseModel
}

type MediaFile struct {
	ID  int64 `json:"id" gorm:"primaryKey"`
	MID int64 `json:"m_id" gorm:"index;not null;comment:所属文件组"`

	FileName   string `json:"file_name" gorm:"not null"`
	Ext        string `json:"ext" gorm:"size:20;comment:扩展名"`
	FileSystem string `json:"file_system" gorm:"size:50;comment:文件系统 local s3 oss"`

	Hash string `json:"hash" gorm:"size:64;uniqueIndex;not null;comment:文件哈希名"`

	Size     int64  `json:"size" gorm:"comment:文件大小"`
	MimeType string `json:"mime_type" gorm:"size:100"`
	DownSec  int64  `json:"down_sec" gorm:"default:0;comment:下载耗费的秒数 "`

	CreatedAt time.Time `json:"create_at" gorm:"index;default:0"`
}

var dbs *db.SQLiteWriter
var getInfoLimit *ratelimit.Limiter

func init() {
	dbs = db.MEDIADB
	dbs.DB().AutoMigrate(&Media{})
	dbs.DB().AutoMigrate(&MediaFile{})
	dbs.DB().AutoMigrate(&MediaUser{})
	dbs.DB().AutoMigrate(&MediaUserTag{})
	dbs.DB().AutoMigrate(&MediaUserToTag{})
	dbs.DB().AutoMigrate(&MediaUserToClient{})
	dbs.DB().AutoMigrate(&MediaUserProxy{})
	dbs.DB().AutoMigrate(&MediaUserDay{})
	dbs.DB().AutoMigrate(&MediaPath{})

	go runstart()

	getInfoLimit = ratelimit.New(
		ratelimit.WithLimit(5, 10*time.Second), // 10秒5次
		ratelimit.WithConcurrency(2),           // 最大2并发
		ratelimit.WithQueue(100),               // 队列长度
	)

	// err := dbs.Write(func(tx *gorm.DB) error {
	// 	return tx.Exec(`
	// 		INSERT INTO media_files (m_id, file_name, file_system, size, mime_type, created_at)
	// 		SELECT
	// 			m.id,
	// 			m.name,
	// 			'local',
	// 			m.size,
	// 			m.mime,
	// 			m.addtime
	// 		FROM media m
	// 		LEFT JOIN media_files f ON f.m_id = m.id
	// 		WHERE f.m_id IS NULL
	// 		`).Error
	// })
	// fmt.Println(err)
}

func (m Media) MarshalJSON() ([]byte, error) {
	type Alias Media
	a := Alias(m)
	if a.Cover != "" {
		a.Cover = storage.Load(a.CoverStorage).URL(a.Cover)
	}

	return config.Json.Marshal(a)
}
func (m MediaFile) MarshalJSON() ([]byte, error) {
	type Alias MediaFile
	a := Alias(m)
	if a.FileName != "" {
		if a.FileSystem == "" {
			a.FileSystem = "local"
		}
		a.FileName = storage.Load(a.FileSystem).URL(a.FileName)
	}

	return config.Json.Marshal(a)
}

func GetDb() *db.SQLiteWriter {
	return dbs
}

// urls 待下载的地址
// pxys 代理
// path 下载到哪个目录
// adminID 管理员id
// systype  使用的存储系统, local 和 minio
// userDir 是否使用用户目录,自动生成平台加用户的独立目录
// redown 是否强制重新下载
func GetPlatformVideos(urls string, pxys []*proxy.ProxyConfig, path string, adminID int64, systype string, userDir, redown bool) []error {
	if systype == "" {
		systype = config.DefStorage
	}
	urlSplit := funcs.ExtractURLs(urls)

	var errs []error
	if len(urlSplit) > 0 {
		mediaMap := getNeedDownUrls(urlSplit, redown)
		var wg sync.WaitGroup
		for _, md := range mediaMap {
			wg.Go(func() {
				t, proxy, pc, _ := getTransport(pxys)
				if pc != nil {
					defer pc.Close(false)
				}
				info, err := parser.ParseVideoShareUrlByRegexp(md.Url, t)
				if err != nil {
					errs = append(errs, errors.New(md.Url+" 下载错误:"+err.Error()))
					fmt.Println("视频解析错误:", err)
					return
				}

				if md.Id < 1 {
					mmd := new(Media)
					dbs.DB().Model(&Media{}).Where("video_id = ?", info.VideoID).Find(mmd)
					if mmd.Id > 0 {
						md = mmd
					}
				}

				// 解析完成后需要先获取平台用户
				var mediaUser *MediaUser
				if info.Author.Uid != "" {
					mu := mediaUserGetter(
						info.Author.Uid,
						info.Author.Name,
						info.Author.Avatar,
						info.Author.SearchID,
						info.Platform,
						proxy,
						adminID,
					)
					mediaUser = mu
					if userDir == true {
						path = fmt.Sprintf("%s/%s/%s", path, info.Platform, info.Author.Name)
					}
				}

				mp, _ := MKDBNameID(path, adminID)
				// fmt.Println(err, "---", mp)
				// panic("----")

				if info.CoverUrl != "" {
					if coverStr, _, _, _, err := storage.Load(systype).Download(mainsignal.MainCtx, info.CoverUrl, &downloader.DownloadOption{
						Proxy: proxy,
					}); err == nil {
						md.Cover = coverStr
						md.CoverStorage = systype
					}
				}

				md.Title = info.Title
				md.VideoID = info.VideoID
				md.UserId = mediaUser.Id
				md.Platform = info.Platform
				md.AdminID = adminID
				md.Sizes = 0
				md.Numbers = 0
				md.User = mediaUser
				if mp != nil {
					md.PathID = mp.ID
				}
				md.Path = path

				var downUrls []string

				if info.VideoUrl != "" {
					downUrls = []string{info.VideoUrl}
				} else if len(info.Images) > 0 {
					for _, v := range info.Images {
						downUrls = append(downUrls, v.Url)
					}
				} else {
					// fmt.Println("找不到任何可下载内容...")
					errs = append(errs, errors.New(md.Url+" 找不到可下载内容!"))
					return
				}

				if err := md.Save(); err != nil {
					errs = append(errs, errors.New(md.Url+" 媒体文件保存失败:"+err.Error()))
					// fmt.Println(err, "media保存错误!", md.Id)
					return
				}

				fls := len(downUrls)

				if fls > 0 {
					var ccv string
					if md.Cover != "" {
						ccv = storage.Load(systype).URL(md.Cover)
					} else {
						ccv = info.CoverUrl
					}
					msg := md.Message(fls, ccv, md.PathID)
					msg.Sent("开始下载...", 0)
					go func() {
						_, err := md.DownMediaFiles(downUrls, proxy, msg)
						if err != nil {
							errs = append(errs, errors.New(md.Url+" 文件下载失败:"+err.Error()))
						}

						var estrs []string
						for _, v := range errs {
							estrs = append(estrs, v.Error())
						}

						err = dbs.Write(func(tx *gorm.DB) error {
							return tx.Exec(`
UPDATE media
SET
    sizes = (
        SELECT SUM(size)
        FROM media_files
        WHERE media_files.m_id = media.id
    ),
    numbers = (
        SELECT COUNT(*)
        FROM media_files
        WHERE media_files.m_id = media.id
    )
WHERE media.id = ?
`, md.Id).Error
						})

						if mediaUser != nil {
							if err := dbs.DB().Model(&Media{}).
								Select("count(*)").
								Where("user_id = ? and removed = 0", mediaUser.Id).
								Scan(&mediaUser.Videos).Error; err == nil {
								dbs.Write(func(tx *gorm.DB) error {
									return tx.Model(&MediaUser{}).Where("id = ?", mediaUser.Id).Update("videos", mediaUser.Videos).Error
								})
							}
						}
						msg.Done(strings.Join(estrs, "\n"))
					}()
				}
			})
			wg.Wait()

		}
	} else {
		errs = append(errs, errors.New("找不到下载连接"))
	}
	return errs
}

// 查找或创建媒体用户
func mediaUserGetter(uuid, name, avatar, nameid, platform, proxy string, adminID int64) *MediaUser {
	mu := new(MediaUser)
	dbs.DB().Model(&MediaUser{}).Where("platform = ? and uuid = ?", platform, uuid).First(mu)

	if mu.Id < 1 {
		mu.Addtime = time.Now().Unix()
		mu.Account = nameid
		mu.Platform = platform
		mu.Uuid = uuid
		mu.Name = name
		mu.Cover = avatar
		mu.AdminID = adminID

		if strings.HasPrefix(strings.ToLower(strings.Trim(avatar, "/")), "www.") {
			avatar = fmt.Sprintf("https://%s", avatar)
		}
		if strings.HasPrefix(avatar, "http") {
			if rsp, _, _, _, err := storage.Load(config.DefStorage).Download(mainsignal.MainCtx, avatar, &downloader.DownloadOption{
				Callback: func(total, downloaded, speed, workers int64) {},
				Proxy:    proxy,
			}); err == nil {
				mu.Cover = rsp
				mu.CoverStorage = config.DefStorage
			}
		}

		if err := dbs.Write(func(tx *gorm.DB) error {
			return tx.Create(mu).Error
		}); err != nil {
			fmt.Println("媒体用户创建失败!")
		}
	}

	return mu
}

func getTransport(pxs []*proxy.ProxyConfig) (*http.Transport, string, *proxy.ProxyConfig, error) {
	var transport *http.Transport
	var proxyStr string
	var pc *proxy.ProxyConfig
	if len(pxs) > 0 {
		pxy := pxs[rand.Intn(len(pxs))]
		if pxy != nil {

			if _, err := pxy.Run(false); err == nil {
				proxyStr = pxy.Listened()
				if proxyURL, err := url.Parse(proxyStr); err == nil {
					transport = &http.Transport{
						Proxy: http.ProxyURL(proxyURL),
					}
				}
			}
			pc = pxy
		}
	}
	return transport, proxyStr, pc, nil
}

// 去重等待下载的url, 返回 md5 => urlstring
func getNeedDownUrls(urls []string, reget bool) map[string]*Media {
	var md5s []string
	md5map := make(map[string]*Media)
	for _, v := range urls {
		md5str := funcs.Md5String(v)
		md5s = append(md5s, md5str)
		md5map[md5str] = &Media{
			Url:    v,
			UrlMd5: md5str,
		}
	}

	var exists []*Media
	dbs.DB().Model(&Media{}).Preload("Files").Preload("User").Where("url_md5 in ?", md5s).Find(&exists)

	for _, v := range exists {
		if reget != true && len(v.Files) > 0 {
			delete(md5map, v.UrlMd5)
		} else if _, ok := md5map[v.UrlMd5]; ok {
			md5map[v.UrlMd5] = v
		}
	}

	return md5map
}

func (m *Media) ReDown(pcs []*proxy.ProxyConfig, adminID int64) (errs []error) {
	t, proxy, pc, _ := getTransport(pcs)
	if pc != nil {
		defer pc.Close(false)
	}

	// 解析地址
	info, err := parser.ParseVideoShareUrlByRegexp(m.Url, t)
	if err != nil {
		errs = append(errs, errors.New(m.Url+" 下载错误:"+err.Error()))
		fmt.Println("视频解析错误:", err)
		return
	}

	// 解析完成后需要先获取平台用户
	var mediaUser *MediaUser
	if info.Author.Uid != "" {
		mu := mediaUserGetter(
			info.Author.Uid,
			info.Author.Name,
			info.Author.Avatar,
			info.Author.SearchID,
			info.Platform,
			proxy,
			adminID,
		)
		mediaUser = mu
	}

	var downUrls []string

	if info.VideoUrl != "" {
		downUrls = []string{info.VideoUrl}
	} else if len(info.Images) > 0 {
		for _, v := range info.Images {
			downUrls = append(downUrls, v.Url)
		}
	} else {
		errs = append(errs, errors.New(m.Url+" 找不到可下载内容!"))
		return
	}

	fls := len(downUrls)
	if fls > 0 {
		var ccv string
		if m.Cover != "" {
			ccv = storage.Load(m.CoverStorage).URL(m.Cover)
		} else {
			ccv = info.CoverUrl
		}
		msg := m.Message(fls, ccv, m.PathID)
		msg.Sent("开始下载...", 0)
		go func() {
			_, err := m.DownMediaFiles(downUrls, proxy, msg)
			if err != nil {
				errs = append(errs, errors.New(m.Url+" 文件下载失败:"+err.Error()))
			}

			var estrs []string
			for _, v := range errs {
				estrs = append(estrs, v.Error())
			}

			err = dbs.Write(func(tx *gorm.DB) error {
				return tx.Exec(`
UPDATE media
SET
    sizes = (
        SELECT SUM(size)
        FROM media_files
        WHERE media_files.m_id = media.id
    ),
    numbers = (
        SELECT COUNT(*)
        FROM media_files
        WHERE media_files.m_id = media.id
    )
WHERE media.id = ?
`, m.Id).Error
			})

			if mediaUser != nil {
				if err := dbs.DB().Model(&Media{}).
					Select("count(*)").
					Where("user_id = ? and removed = 0", mediaUser.Id).
					Scan(&mediaUser.Videos).Error; err == nil {
					dbs.Write(func(tx *gorm.DB) error {
						return tx.Model(&MediaUser{}).Where("id = ?", mediaUser.Id).Update("videos", mediaUser.Videos).Error
					})
				}
			}
			msg.Done(strings.Join(estrs, "\n"))
		}()
	}
	return
}

func (m *Media) DownMediaFiles(fls []string, proxy string, msg *MediaResponseMessage) ([]*MediaFile, error) {
	var dbfls []*MediaFile
	var fdbfls []*MediaFile
	var doned int
	for _, v := range fls {
		fn, sz, downloadSec, mime, err := storage.Load("").Download(
			mainsignal.MainCtx,
			v,
			&downloader.DownloadOption{
				Proxy: proxy,
				Callback: func(total, downloaded, speed, workers int64) {
					msgstr := fmt.Sprintf(
						"%.2f%% %s/s %s",
						float64(downloaded)/float64(total)*100,
						funcs.FormatFileSize(speed, "1", ""),
						funcs.FormatFileSize(total, "1", ""),
					)
					if msg != nil {
						msg.Sent(msgstr, doned)
					} else {
						fmt.Printf("\r%s", msgstr)
					}
				},
			},
		)
		doned++
		msg.Sent(fmt.Sprintf("已下载:%d", doned), doned)
		if err == nil {
			nsps := strings.Split(filepath.Base(fn), ".")
			hash := nsps[0]

			var mdrow MediaFile
			dbs.DB().Model(&MediaFile{}).Where("hash = ?", hash).First(&mdrow)
			if mdrow.ID < 1 {
				dbfls = append(dbfls, &MediaFile{
					MID:        m.Id,
					FileName:   fn,
					FileSystem: config.DefStorage,
					Ext:        strings.Trim(filepath.Ext(fn), "."),
					Hash:       hash,
					MimeType:   mime,
					Size:       sz,
					DownSec:    downloadSec,
					CreatedAt:  time.Now(),
				})
			} else {
				fdbfls = append(fdbfls, &mdrow)
			}

		} else {
			fmt.Println("下载错误:", err)
		}
	}
	if len(dbfls) > 0 {
		if err := dbs.Write(func(tx *gorm.DB) error {
			return tx.Create(dbfls).Error
		}); err != nil {
			return nil, err
		}
	}
	return append(fdbfls, dbfls...), nil
}

func (m *Media) Save() error {
	return dbs.Write(func(tx *gorm.DB) error {
		if m.Id > 0 {
			if err := tx.Model(&Media{}).Where("id = ?", m.Id).Updates(map[string]any{
				"path_id":  m.PathID,
				"path":     m.Path,
				"title":    m.Title,
				"platform": m.Platform,
				"removed":  m.Removed,
				"url":      m.Url,
				"url_md5":  m.UrlMd5,
				"user_id":  m.UserId,
				"video_id": m.VideoID,
				"sizes":    m.Sizes,
				"numbers":  m.Numbers,
				"admin_id": m.AdminID,
				"cover":    m.Cover,
			}).Error; err != nil {
				fmt.Println("media 更新失败:", err)
				return err
			}
		} else {
			m.Addtime = time.Now()
			return tx.Create(m).Error
		}
		return nil
	})
}

func Downloads() {

}

// 以下方法即将废弃
func MkerMediaUser(platform, uid, cover, name, proxy, searchID string, adminID int64) *MediaUser {
	mu := new(MediaUser)
	if err := dbs.DB().Model(&MediaUser{}).Where("platform = ? and uuid = ?", platform, uid).First(mu).Error; err != nil {
		mu.Cover = cover

		if name, err := downloader.Download(mainsignal.MainCtx, &downloader.DownloadOption{
			URL:      cover,
			Dir:      config.FullPath(config.MEDIAROOT, "avatar"),
			FileName: funcs.Md5String(cover),
		}); err == nil {
			mu.Cover = name.FullName
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

// func GetMediasUserFromName(names []string) map[string]*MediaUser {
// 	var mmus []*Media

// 	resp := make(map[string]*MediaUser)
// 	if err := dbs.DB().Model(&Media{}).Where("name in ?", names).Find(&mmus).Error; err == nil {
// 		var ids []int64
// 		// idNames := make(map[int64]string)
// 		for _, v := range mmus {
// 			if v.UserId > 0 {
// 				ids = append(ids, v.UserId)
// 				// idNames[v.UserId] = v.Name
// 			}
// 		}

// 		var mmuus []*MediaUser
// 		if len(ids) > 0 {
// 			dbs.DB().Model(&MediaUser{}).Where("id in ?", ids).Find(&mmuus)
// 			sdsd := make(map[int64]*MediaUser)
// 			for _, v := range mmuus {
// 				sdsd[v.Id] = v
// 			}

// 			for _, v := range mmus {
// 				if vvu, ok := sdsd[v.UserId]; ok {
// 					resp[v.Name] = vvu
// 				}
// 			}
// 		}
// 	}
// 	return resp
// }

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

		// for _, m := range ms {
		// 	origin := config.FullPath(filepath.Join(config.MEDIAROOT, m.Path, m.Name))
		// 	if _, err := os.Stat(origin); err != nil {
		// 		continue
		// 	}
		// 	if err := os.Rename(
		// 		origin,
		// 		config.FullPath(config.TRASHED, m.Name),
		// 	); err != nil {
		// 		return err
		// 	}
		// }

		return nil
	})
}
