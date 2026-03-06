package users

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/medias"
	"tools/runtimes/db/messages"
	"tools/runtimes/db/proxys"
	"tools/runtimes/db/task"
	"tools/runtimes/downloader"
	"tools/runtimes/funcs"
	"tools/runtimes/listens/ws"
	"tools/runtimes/mainsignal"
	"tools/runtimes/response"
	"tools/runtimes/runner"
	"tools/runtimes/videos/downloader/parser"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

var BathTask *task.Task

func init() {
	if t, err := task.NewTask("batchuseradd", 0, "批量添加用户", 3, true); err == nil {
		BathTask = t
	}
}

// 获取用户
func List(c *gin.Context) {
	admin, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请登录", nil)
		return
	}

	dt := new(db.ListFinder)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	list, total := medias.GetMediaUsers(admin.Id, dt)
	rsp := gin.H{
		"list":  list,
		"total": total,
	}

	response.Success(c, rsp, "")
}

func GetTags(c *gin.Context) {
	response.Success(c, medias.GetTags(), "")
}

func GetPlatforms(c *gin.Context) {
	response.Success(c, medias.GetUserPlatforms(), "")
}

func Editer(c *gin.Context) {
	admin, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请登录", nil)
		return
	}

	mu := new(medias.MediaUser)
	if err := c.ShouldBind(mu); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if mu.Id < 1 {
		response.Error(c, http.StatusBadRequest, "错误的请求:id", nil)
		return
	}

	// MediaUser不能让改其他数据,因此限制了可编辑的字段
	mmu := medias.GetMediaUserByID(mu.Id)
	if mmu == nil || mmu.Id < 1 {
		response.Error(c, http.StatusBadRequest, "错误的请求:nil", nil)
		return
	}
	// 其他人不允许编辑
	if mmu.AdminID != admin.Id {
		response.Error(c, http.StatusBadRequest, "错误的请求:uid", nil)
		return
	}

	if err := medias.GetDb().Write(func(tx *gorm.DB) error {
		if err := mmu.EmptyClient(tx); err != nil {
			return err
		}
		if err := mmu.EmptyProxy(tx); err != nil {
			return err
		}
		if err := mmu.EmptyTag(tx); err != nil {
			return err
		}

		for tp, vls := range mu.Clients {
			var mutc []*medias.MediaUserToClient
			for _, v := range vls {
				mutc = append(mutc, &medias.MediaUserToClient{
					MUID:       mu.Id,
					ClientType: tp,
					ClientID:   v,
				})
			}
			if len(mutc) > 0 {
				if err := tx.Create(mutc).Error; err != nil {
					return err
				}
			}
		}

		if len(mu.Proxys) > 0 {
			var mutp []*medias.MediaUserProxy
			for _, v := range mu.Proxys {
				mutp = append(mutp, &medias.MediaUserProxy{
					MUID:    mu.Id,
					ProxyID: v,
				})
			}
			if len(mutp) > 0 {
				if err := tx.Create(mutp).Error; err != nil {
					return err
				}
			}
		}

		if len(mu.Tags) > 0 {
			tgs := medias.AddMUTagsBySlice(mu.Tags)
			var mutt []*medias.MediaUserToTag
			for _, v := range tgs {
				mutt = append(mutt, &medias.MediaUserToTag{
					UserID: mu.Id,
					TagID:  v.ID,
				})
			}
			if len(mutt) > 0 {
				if err := tx.Create(mutt).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	// fmt.Println("其他信息处理完成... 保存")

	mmu.Autoinfo = mu.Autoinfo
	mmu.AutoDownload = mu.AutoDownload
	mmu.AutoTimer = mu.AutoTimer
	mmu.DownFreq = mu.DownFreq
	if err := mmu.Save(mmu, medias.GetDb().DB()); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	go mmu.AutoStart()
	mu.Commpare()

	response.Success(c, mu, "")
}

func Delete(c *gin.Context) {

}

type BatchAddReq struct {
	Urls   []string `json:"urls"`
	Proxy  int64    `json:"proxy"`
	Client int64    `json:"client"`
}

var runidx int64

func BatchAdd(c *gin.Context) {
	u, err := admins.GetAdminUser(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	req := new(BatchAddReq)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	total := len(req.Urls)

	if BathTask != nil {
		runidx++
		tr, err := BathTask.AddInterval(
			fmt.Sprintf("%d", runidx),
			time.Now().Format("2006-01-02 15:04:05"),
			0,
			time.Hour*6,
			3,
			time.Second*5,
			time.Time{},
			func(tr *task.TaskRun) error {
				go mkssr(req, u.Id, tr)
				return nil
			},
		)
		if err == nil {
			tr.Total = float64(total)
			tr.RunNow()
		} else {
			fmt.Println(err, "---tr error")
		}
	}
	// go mkssr(req, u.Id)
	response.Success(c, nil, fmt.Sprintf("%d 条数据已添加后台自动获取", total))
}

func mkssr(req *BatchAddReq, adminid int64, tr *task.TaskRun) {
	tr.Total = float64(len(req.Urls))
	// 构造客户端和代理
	opt := browserdb.GenBrowserOpt(req.Client, false)
	if opt.Pc == nil && req.Proxy > 0 {
		pp := proxys.GetById(req.Proxy)
		if pp != nil {
			go func() {
				if ppp, err := pp.Start(false); err == nil {
					opt.Pc = ppp
				}
			}()
		}
	}

	var proxy string
	if opt != nil && opt.Pc != nil {
		defer opt.Pc.Close(false)
		proxy = opt.Pc.Listened()
	}
	doned := 0
	errored := 0
	jscode := "douyin-info"
	js := jses.GetJsByCode(jscode)
	if js != nil && js.ID > 0 {
		opt.JsStr = js.GetContent(nil)
		var vids []string
		var uids []string
		videoEndmap := make(map[string]string)
		userEndmap := make(map[string]string)
		for _, v := range req.Urls { //8.76 o@D.Hi mQx:/ 11/05 爱的桥段让我怎么写# 阳光是最好的滤镜 # 营业  https://v.douyin.com/ZPZ6xisL_pk/ 复制此链接，打开Dou音搜索，直接观看视频！
			if !strings.Contains(v, ".douyin.com") {
				errored++

				tr.Doned = tr.Doned + 1
				tr.SentMsg(fmt.Sprintf("%s 不被支持", v), 0, false)
			} else {
				v = strings.Trim(v, "/")
				vs := strings.Split(v, "?")
				v = vs[0]

				vvs := strings.Split(v, "/")
				endstr := vvs[len(vvs)-1]

				if strings.Contains(v, "/user/") {
					uids = append(uids, endstr)
					userEndmap[endstr] = v
				} else {
					if strings.Contains(v, "/video/") {
						vids = append(vids, endstr)
					}
					videoEndmap[endstr] = v
				}
			}
		}

		wg := new(sync.WaitGroup)
		if len(uids) > 0 {
			var mus []*medias.MediaUser
			medias.GetDb().DB().Model(&medias.MediaUser{}).Where("uuid in ?", uids).Find(&mus)
			for _, v := range mus {
				if urlstr, ok := userEndmap[v.Uuid]; ok {
					errored++
					delete(userEndmap, v.Uuid)
					tr.Doned = tr.Doned + 1
					tr.SentMsg(fmt.Sprintf("%s 已添加", urlstr), 0, false)
				}
			}
			wg.Go(func() {
				select {
				case <-mainsignal.MainCtx.Done():
					return
				default:
					for uuid, v := range userEndmap {
						opt.Url = v
						r, err := runner.GetRunner(0, opt)
						if err == nil {
							r.Start(time.Second*60, func(s string) error {
								gs := gjson.Parse(s)

								mu := new(medias.MediaUser)
								mu.Addtime = time.Now().Unix()
								mu.AdminID = adminid
								mu.Uuid = uuid
								mu.Name = gs.Get("name").String()
								mu.Cover = gs.Get("cover").String()
								mu.Works = gs.Get("works").Int()
								mu.Fans = gs.Get("fans").Int()
								mu.Local = gs.Get("local").String()
								mu.Account = gs.Get("account").String()
								mu.Sex = gs.Get("sex").String()
								mu.Platform = "douyin"

								if mu.Cover != "" {
									exts := "png"
									dl := downloader.NewDownloader(proxy, nil, nil)
									if ext, err := dl.GetUrlFileExt(mu.Cover); err == nil {
										exts = ext
									}
									dest := fmt.Sprintf("avatar/%s.%s", funcs.Md5String(mu.Cover), exts)
									fullPath := filepath.Join(config.MEDIAROOT, dest)
									dirRoot := filepath.Dir(fullPath)
									if _, err := os.Stat(dirRoot); err != nil {
										os.MkdirAll(dirRoot, os.ModePerm)
									}

									if err := dl.Download(mu.Cover, fullPath); err == nil {
										mu.Cover = dest
									}

									if err := mu.Save(mu, medias.GetDb().DB()); err != nil {
										return err
									}
									doned++

									return nil
								}
								fmt.Println("返回的数据: ", s)
								errored++
								return errors.New("找不到数据!")
							})
						} else {
							errored++
						}
						tr.Doned = tr.Doned + 1
						tr.SentMsg(fmt.Sprintf("总数: %d, 成功: %d, 失败: %d %s", len(req.Urls), doned, errored, err.Error()), 0, false)
						time.Sleep(time.Second * time.Duration(funcs.RandomNumber(3, 8)))
					}
				}

			})
		}

		if len(vids) > 0 {
			var vmedias []*medias.Media
			medias.GetDb().DB().Model(&medias.Media{}).Where("video_id in ?", vids).Find(&vmedias)
			for _, v := range vmedias {
				if urlstr, ok := videoEndmap[v.VideoID]; ok {
					errored++
					delete(videoEndmap, v.VideoID)

					tr.Doned = tr.Doned + 1
					tr.SentMsg(fmt.Sprintf("%s 已添加", urlstr), 0, false)
				}
			}

			wg.Go(func() {
				var transport *http.Transport
				if proxy != "" {
					if proxyURL, err := url.Parse(proxy); err == nil {
						transport = &http.Transport{
							Proxy: http.ProxyURL(proxyURL),
						}
					}
				}

				select {
				case <-mainsignal.MainCtx.Done():
					return
				default:
					for _, v := range videoEndmap {
						parseRes, err := parser.ParseVideoShareUrl(v, transport)
						if err != nil {
							errored++
						} else {
							if parseRes.Author.Uid != "" {
								mu := medias.MkerMediaUser(parseRes.Platform, parseRes.Author.Uid, parseRes.Author.Avatar, parseRes.Author.Name, proxy, parseRes.Author.SearchID, adminid)
								savePath := mu.DefDirName(parseRes.Author.Uid) //fmt.Sprintf(".auto/%s", funcs.Md5String(parseRes.Author.Uid))

								if parseRes.VideoUrl != "" {
									md, err := medias.DownLoadVideo(v, parseRes.VideoUrl, savePath, "", proxy, func(percent float64, downloaded, total int64) {
										fmt.Printf("\r下载进度: %.2f%%", percent)
									})

									if err != nil {
										errored++
									} else {
										md.VideoID = parseRes.VideoID
										md.Platform = parseRes.Platform
										md.Title = parseRes.Title
										md.UserId = mu.Id
										if err := md.Save(md, medias.GetDb().DB()); err != nil {
											errored++
										} else {
											doned++
										}
									}
								} else {
									errored++
									err = errors.New("不是视频")
								}
							}
						}

						tr.Doned = tr.Doned + 1
						var errmsg string
						if err != nil {
							errmsg = err.Error()
						}
						tr.SentMsg(fmt.Sprintf("总数: %d, 成功: %d, 失败: %d %s", len(req.Urls), doned, errored, errmsg), 0, false)
						time.Sleep(time.Second * time.Duration(funcs.RandomNumber(4, 9)))
					}
				}
			})
		}

		wg.Wait()

		messages.SuccessMsg(fmt.Sprintf("总数: %d, 成功: %d, 失败: %d", len(req.Urls), doned, errored))
		if bt, err := config.Json.Marshal(map[string]any{
			"type": "batchadddone",
			"data": "",
		}); err == nil {
			ws.Broadcost(bt)
		}
	} else {
		messages.ErrorMsg("找不到执行js")
	}
	tr.RemoveMsg()
	if BathTask != nil {
		BathTask.Stop(tr.RunID)
	}
}

// 查看本地视频
type localVideoObj struct {
	Page  int     `json:"page"`
	Limit int     `json:"limit"`
	Q     string  `json:"q"`
	UIDs  []int64 `json:"uids"`
}

func LocalVideo(c *gin.Context) {
	lob := new(localVideoObj)
	if err := c.ShouldBindJSON(lob); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if lob.Page < 1 {
		lob.Page = 1
	}
	if lob.Limit < 1 {
		lob.Limit = 20
	}

	type aaab struct {
		Mid        int64  `json:"mid"`
		Muid       int64  `json:"muid"`
		Size       int64  `json:"size"`
		Filename   string `json:"filename"`
		Platform   string `json:"platform"`
		Filetime   int64  `json:"filetime"`
		Path       string `json:"path"`
		Mime       string `json:"mime"`
		MediaTitle string `json:"media_titme"`
		UserTitle  string `json:"user_title"`
		UserCover  string `json:"user_cover"`
		VideoID    string `json:"video_id"`
		Uuid       string `json:"uuid"`
		Origin     string `json:"origin"`
		Local      string `json:"local"`
		Fans       int64  `json:"fans"`
		Works      int64  `json:"works"`
		Downloaded int64  `json:"downloaded"`
	}

	var lists []*aaab
	var total int64
	md := medias.GetDb().DB().
		Model(&medias.Media{}).
		Select(
			"media.id as mid",
			"media.user_id as muid",
			"media.size",
			"media.name as filename",
			"media.platform",
			"media.filetime",
			"media.path",
			"media.mime",
			"media.title as media_title",
			"mu.name as user_title",
			"mu.cover as user_cover",
			"media.video_id",
			"mu.uuid",
			"media.url as origin",
			"mu.works",
			"mu.fans",
			"mu.local",
			"mu.videos as downloaded",
		).
		Joins("left join media_users as mu on mu.id = media.user_id").
		Where("media.trashed = 0")
	if len(lob.UIDs) > 0 {
		md = md.Where("user_id in ?", lob.UIDs)
	}
	if lob.Q != "" {
		md = md.Where("media.title like ? ESCAPE '\\'", "%"+lob.Q+"%")
	}
	md.Count(&total)
	md.Offset((lob.Page - 1) * lob.Limit).Limit(lob.Limit).Order("media.id DESC").Scan(&lists)

	var listmap []map[string]any
	for _, v := range lists {
		listmap = append(listmap, map[string]any{
			"size":     funcs.FormatFileSize(v.Size),
			"url":      fmt.Sprintf("%s/%s/%s", config.MediaUrl, v.Path, v.Filename),
			"title":    v.MediaTitle,
			"id":       v.Mid,
			"origin":   v.Origin,
			"platform": v.Platform,
			"filetime": v.Filetime,
			"mime":     v.Mime,
			"path":     v.Path,
			"user": map[string]any{
				"id":         v.Muid,
				"cover":      fmt.Sprintf("%s/%s", config.MediaUrl, v.UserCover),
				"title":      v.UserTitle,
				"uuid":       v.Uuid,
				"works":      v.Works,
				"fans":       v.Fans,
				"local":      v.Local,
				"downloaded": v.Downloaded,
			},
		})
	}
	response.Success(c, gin.H{
		"total": total,
		"lists": listmap,
		"pages": (total + int64(lob.Limit) - 1) / int64(lob.Limit),
	}, "")
}
