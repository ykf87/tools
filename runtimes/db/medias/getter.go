package medias

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/config"
	"tools/runtimes/db/jses"
	"tools/runtimes/eventbus"
	"tools/runtimes/funcs"
	"tools/runtimes/listens/ws"
	"tools/runtimes/videos/downloader/parser"

	"github.com/chromedp/cdproto/runtime"
	"github.com/tidwall/gjson"
)

// 用户每天的粉丝作品等数据
type MediaUserDay struct {
	MUID  int64  `json:"m_uid" gorm:"primaryKey;not null;"` // media_users表的id
	Ymd   string `json:"ymd" gorm:"primaryKey;not null"`    // 年月日的时间格式
	Works int64  `json:"works" gorm:"index;default:-1"`     // 发布作品数量
	Fans  int64  `json:"fans" gorm:"index;default:-1"`      // 粉丝数
	Zan   int64  `json:"zan" gorm:"index;default:-1"`       // 获赞数
}

var autoLoaderUser sync.Map
var today string
var mugetterBs *bs.Manager

func init() {
	mugetterBs = bs.NewManager("")
}

func autoStart() {
	for {
		autoLoaderUser.Range(func(k, v any) bool {
			if mu, ok := v.(*MediaUser); ok {
				if mu.Autoinfo == 0 && mu.AutoDownload == 0 {
					autoLoaderUser.Delete(k)
					return true
				}
				if mu.Isruner == false {
					go mu.runner()
				}
			}
			return true
		})
		time.Sleep(time.Second * 30)
	}
}

func (t *MediaUser) runner() {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Println(t.Id, "------")
	// 如果存在自动下载
	if t.AutoDownload == 1 {

	}

	t.done = make(chan bool)
	bs, err := t.StartGetter(
		func(str string, bs *bs.Browser) {
			dt := gjson.Parse(str)
			if fans := dt.Get("fans").Int(); fans > 0 {
				t.Fans = fans
			}
			if works := dt.Get("works").Int(); works > 0 {
				t.Works = works
			}
			if local := dt.Get("local").String(); local != "" {
				t.Local = local
			}
			if account := dt.Get("account").String(); account != "" {
				t.Account = account
			}
			if t.AutoDownload == 1 && dt.Get("lists").Exists() {
				var vids []string
				for _, v := range dt.Get("lists").Array() {
					vids = append(vids, v.String())
				}
				fmt.Println("找到带下载视频列表:", vids)
				t.autodownload(vids)
			}
			t.LastDownTime = time.Now().Unix()
			t.Save(nil)
			t.Commpare()
			bs.Close()
			close(t.done)
		},
		func() {
			eventbus.Bus.Publish("media_user_info", t)
		},
		func(url string) {

		},
	)
	if err == nil {
		t.Isruner = true
		go func() {
			sleep := 30 * time.Second
			timer := time.NewTimer(sleep)
			defer timer.Stop()
			for {
				select {
				case <-timer.C:
					if bs.IsArrive() {
						bs.Close()
						return
					}
					timer.Reset(sleep)
				case <-t.done:
					return
				}
			}
		}()
	}
}

func (t *MediaUser) StartGetter(consoleFun func(str string, bs *bs.Browser), closeFun func(), urlchangeFun func(url string)) (*bs.Browser, error) {
	var jscode string
	var runurl string
	switch t.Platform {
	case "douyin":
		jscode = "douyin-info"
		runurl = fmt.Sprintf("https://www.douyin.com/user/%s", t.Uuid)
	default:
		return nil, fmt.Errorf("暂时不支持 %s 获取获取信息", t.Platform)
	}

	js := jses.GetJsByCode(jscode)
	if js == nil || js.ID < 1 {
		return nil, fmt.Errorf("获取账号信息脚本不存在")
	}
	runjs := js.GetContent(nil)

	brows, _ := mugetterBs.New(0, bs.Options{
		Url:      runurl,
		JsStr:    runjs,
		Headless: false,
		Timeout:  time.Duration(time.Second * 30),
	})

	brows.OnClosed(func() {
		if closeFun != nil {
			closeFun()
		}
	})

	if consoleFun != nil {
		brows.OnConsole(func(args []*runtime.RemoteObject) {
			for _, arg := range args {
				if arg.Value != nil {
					gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
					if gs.Get("type").String() == "kaka" {
						consoleFun(gs.Get("data").String(), brows)
					}
				}
			}
		})
	}

	if urlchangeFun != nil {
		brows.OnURLChange(func(url string) {
			urlchangeFun(url)
		})
	}

	if err := brows.OpenBrowser(); err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 1)
	go brows.RunJs(runjs)

	return brows, nil
}

// 自动下载短视频
type pms struct {
	Name       string     `json:"name"`         // 文件名称
	Path       string     `json:"path"`         // 路径名称
	FullName   string     `json:"full_name"`    // 运行目录下的相对路径
	Url        string     `json:"url"`          // 链接地址
	Timer      int64      `json:"timer"`        // 最后更新时间
	Dir        bool       `json:"dir"`          // 是否是目录
	Ext        string     `json:"ext"`          // 文件后缀
	Size       string     `json:"size"`         // 文件大小
	Mime       string     `json:"mime"`         // 文件类型
	Fmt        string     `json:"fmt"`          // 下载百分比字符串
	Num        float64    `json:"num"`          // 下载进度数字,100为下载完成
	DownFile   string     `json:"down_file"`    // md5内容,下载地址的md5
	Status     int        `json:"status"`       // 下载状态
	DownErrMsg string     `json:"down_err_msg"` // 下载错误信息
	Platform   string     `json:"platform"`     // 下载的平台
	Cover      string     `json:"cover"`        // 封面
	User       *MediaUser `json:"user" gorm:"-"`
}

func (t *MediaUser) autodownload(videos []string) {
	var lastGetVideoId string
	dbs.Model(&Media{}).Select("video_id").Where("user_id = ?", t.Id).Order("id DESC").First(&lastGetVideoId)

	maxget := 4
	var waitdown []string
	for idx, v := range videos {
		if idx >= maxget {
			break
		}
		if lastGetVideoId != "" && v == lastGetVideoId {
			break
		}
		waitdown = append(waitdown, v)
	}

	if len(waitdown) < 1 {
		return
	}

	// 如果有
	wg := new(sync.WaitGroup)
	for _, v := range waitdown {
		wg.Go(func() {
			var transport *http.Transport
			if t.trans != "" {
				if proxyURL, err := url.Parse(t.trans); err == nil {
					transport = &http.Transport{
						Proxy: http.ProxyURL(proxyURL),
					}
				}
			}

			parseRes, err := parser.ParseVideoShareUrlByRegexp(v, transport)
			if err != nil {
				return
			}
			if parseRes.VideoUrl != "" {
				fn := funcs.Md5String(parseRes.VideoUrl)
				path := fmt.Sprintf(".auto/%s%d", t.Uuid, t.Id)
				md, err := DownLoadVideo(parseRes.VideoUrl, path, "", t.trans, func(percent float64, downloaded, total int64) {
					fmt.Printf("\r下载进度: %.2f%%", percent)
					dbk := new(pms)
					dbk.DownFile = fn
					dbk.Fmt = fmt.Sprintf("%.2f%%", percent)
					dbk.Num = percent
					dbk.Dir = false
					dbk.Cover = parseRes.CoverUrl
					dbk.Name = fn
					dbk.Platform = parseRes.Platform

					ws.SentBus(t.AdminID, "video-download", dbk, "")
				})

				dbk := new(pms)
				if err != nil {
					dbk.DownFile = fn
					dbk.Fmt = ""
					dbk.Num = 0
					dbk.Dir = false
					dbk.Status = -1
					dbk.DownErrMsg = err.Error()

					ws.SentBus(t.Id, "video-download", dbk, "")
					return
				}
				md.Title = parseRes.Title
				md.Platform = parseRes.Platform
				md.VideoID = parseRes.VideoID
				md.UserId = t.Id
				if err := md.Save(nil); err != nil {
					return
				}

				dbk.DownFile = fn
				dbk.Fmt = "100%"
				dbk.Num = 100
				dbk.Dir = false
				dbk.Status = 1
				dbk.Mime = md.Mime
				dbk.Size = funcs.FormatFileSize(md.Size)
				dbk.Name = md.Name
				dbk.Platform = md.Platform
				dbk.Url = fmt.Sprintf("%s/%s", config.MediaUrl, filepath.Join(path, md.Name))

				rrs := GetMediasUserFromName([]string{md.Mime})
				if vvs, ok := rrs[md.Mime]; ok {
					dbk.User = vvs
				}

				tms := strings.Split(md.Mime, ".")
				if len(tms) > 1 {
					dbk.Ext = strings.ToLower(tms[len(tms)-1])
				}

				dbk.FullName = fmt.Sprintf("%s/%s", md.Path, md.Name)
				dbk.Timer = md.Filetime

				ws.SentBus(t.AdminID, "video-download", dbk, "")
			}
		})
	}
	wg.Wait()
}
