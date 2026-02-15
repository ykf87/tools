package medias

import (
	"context"
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
	"tools/runtimes/db/task"
	"tools/runtimes/funcs"
	"tools/runtimes/listens/ws"
	"tools/runtimes/logs"
	"tools/runtimes/scheduler"
	"tools/runtimes/videos/downloader/parser"

	"github.com/chromedp/cdproto/runtime"
	"github.com/tidwall/gjson"
)

type AutoInfo struct {
	MID int64

	ctx    context.Context
	cancle context.CancelFunc
	runner *scheduler.Runner
}

type AutoDown struct {
	MID    int64
	ctx    context.Context
	cancle context.CancelFunc
	runner *scheduler.Runner
	bs     *bs.Browser
}

var autodown sync.Map // 自动下载
var autoinfo sync.Map // 自动获取信息
var today string

// var downsch *scheduler.Scheduler // 自动下载调度器
// var infosch *scheduler.Scheduler // 自动获取信息调度器

// var TaskLogger *tasklog.Task

var autoDownTL *task.Task
var autoInfoTL *task.Task

func runstart() {
	// downsch = scheduler.NewWithLimit(mainsignal.MainCtx, 10)
	// infosch = scheduler.NewWithLimit(mainsignal.MainCtx, 5)

	// TaskLogger = tasklog.NewTaskLog("autoupdatemediauser", "自动更新/下载用户信息", 10, 30, false)
	if ad, err := task.NewTask("mediauserdown", 0, "用户自动下载任务", 10, false); err == nil {
		autoDownTL = ad
	}
	if ad, err := task.NewTask("mediauserinfo", 0, "用户自动更新任务", 10, false); err == nil {
		autoInfoTL = ad
	}
	for _, v := range GetAutoUsers() {
		v.AutoStart()
	}
}

func (mu *MediaUser) DownID() string {
	return fmt.Sprintf("muautodown-%d", mu.Id)
}

func (mu *MediaUser) InfoID() string {
	return fmt.Sprintf("muautoinfo-%d", mu.Id)
}

func (mu *MediaUser) AutoStart() error {
	if opt, err := getRunnerOption(mu.Id); err == nil {
		opt.Stop()
	}
	if mu.AutoDownload == 1 && mu.DownFreq > 0 {
		did := mu.DownID()
		tr, err := autoDownTL.AddChild(did, fmt.Sprintf("%s 自行下载中...", mu.Name), time.Second*600)
		if err != nil {
			logs.Error(err.Error())
		} else {
			err := tr.StartInterval(int64(mu.DownFreq*60), func(tr *task.TaskRun) error {
				fmt.Println("自动下载中----")
				time.Sleep(time.Second * 2)
				// mu.StartGetter(func(str string, bs *bs.Browser) {

				// }, func() {

				// }, func(url string) {

				// })
				return nil
			})
			fmt.Println(err, "自动下载错误------")
		}

		// if tsk := TaskLogger.GetRunner(did); tsk == nil {
		// 	opt, err := GetOptions(mu)
		// 	if err != nil {
		// 		logs.Error(err.Error())
		// 		return err
		// 	}

		// 	opt.runner = TaskLogger.Append(
		// 		mainsignal.MainCtx,
		// 		did,
		// 		fmt.Sprintf("%s 自动下载", mu.Name),
		// 		opt.FmtDownload,
		// 	)
		// 	if err := opt.Start(); err != nil {
		// 		logs.Error(err.Error())
		// 		return err
		// 	}
		// }
	}
	if mu.Autoinfo == 1 {
		iID := mu.InfoID()
		tr, err := autoInfoTL.AddChild(iID, fmt.Sprintf("%s 自动获取信息", mu.Name), time.Second*30)
		if err != nil {
			logs.Error(err.Error())
		} else {
			tr.StartAtTime(mu.AutoTimer, func(tr *task.TaskRun) error {
				fmt.Println("自动更新信息...")
				return nil
			})
		}
		// h, m, s := funcs.MsToHMS(mu.AutoTimer)
		// if tsk := TaskLogger.GetRunner(iID); tsk == nil {
		// 	opt, err := GetOptions(mu)
		// 	if err != nil {
		// 		logs.Error(err.Error())
		// 		return err
		// 	}
		// 	opt.runner = TaskLogger.Append(
		// 		mainsignal.MainCtx,
		// 		iID,
		// 		fmt.Sprintf("%s 自动更新用户数据", mu.Name),
		// 		func(msg string, tr *tasklog.TaskRunner) error {
		// 			return mu.ParseUserInfoData(msg)
		// 		},
		// 	)
		// 	if err := opt.StartDailyRandomAt(h, m, s, 20); err != nil {
		// 		logs.Error(err.Error())
		// 		return err
		// 	}
		// }
	}
	return nil
}

func (mu *MediaUser) AutoStarts() {
	// if mu.AutoDownload == 1 && mu.DownFreq > 0 {

	// 	fmt.Println("自动下载:", mu.Id)
	// 	if _, ok := autodown.Load(mu.Id); ok {
	// 		return
	// 	}
	// 	ad := &AutoDown{
	// 		MID: mu.Id,
	// 	}
	// 	ad.ctx, ad.cancle = context.WithCancel(mainsignal.MainCtx)
	// 	runner := downsch.NewRunner(func(ctx context.Context) error {
	// 		bs, err := mu.StartGetter(
	// 			func(str string, bs *bs.Browser) {
	// 				dt := gjson.Parse(str)
	// 				if fans := dt.Get("fans").Int(); fans > 0 {
	// 					mu.Fans = fans
	// 				}
	// 				if works := dt.Get("works").Int(); works > 0 {
	// 					mu.Works = works
	// 				}
	// 				if local := dt.Get("local").String(); local != "" {
	// 					mu.Local = local
	// 				}
	// 				if account := dt.Get("account").String(); account != "" {
	// 					mu.Account = account
	// 				}
	// 				if dt.Get("lists").Exists() {
	// 					var vids []string
	// 					for _, v := range dt.Get("lists").Array() {
	// 						vids = append(vids, v.String())
	// 					}
	// 					fmt.Println("找到带下载视频列表:", vids)
	// 					mu.autodownload(vids)
	// 				}
	// 				mu.LastDownTime = time.Now().Unix()
	// 				dbs.Write(func(tx *gorm.DB) error {
	// 					return mu.Save(mu, tx)
	// 				})

	// 				mu.Commpare()
	// 				bs.Close()
	// 			},
	// 			func() {
	// 				eventbus.Bus.Publish("media_user_info", mu)
	// 			},
	// 			func(url string) {

	// 			},
	// 		)
	// 		if err != nil {
	// 			return nil
	// 		}
	// 		ad.bs = bs
	// 		return nil
	// 	}, time.Second*120, ad.ctx)

	// 	runner.Every(time.Duration(mu.DownFreq) * time.Minute).SetCloser(func() {
	// 		fmt.Println("关闭自动下载任务")
	// 	}).SetError(func(err error, tried int32) {
	// 		fmt.Println("自动下载错误:", err)
	// 	}).RunNow()
	// }

	// if mu.Autoinfo == 1 && mu.AutoTimer > 0 {
	// 	if _, ok := autoinfo.Load(mu.Id); ok {
	// 		return
	// 	}
	// }

	// return
	// for {
	// 	autoLoaderUser.Range(func(k, v any) bool {
	// 		if mu, ok := v.(*MediaUser); ok {
	// 			if mu.Autoinfo == 0 && mu.AutoDownload == 0 {
	// 				autoLoaderUser.Delete(k)
	// 				return true
	// 			}
	// 			if mu.Isruner == false {
	// 				go mu.runner()
	// 			}
	// 		}
	// 		return true
	// 	})
	// 	time.Sleep(time.Second * 30)
	// }
}

func (mu *MediaUser) Stop() {
	if v, ok := autodown.Load(mu.Id); ok {
		if vv, ok := v.(*AutoDown); ok {
			vv.cancle()
		}
		autodown.Delete(mu.Id)
	}
	if v, ok := autoinfo.Load(mu.Id); ok {
		if vv, ok := v.(*AutoInfo); ok {
			vv.cancle()
		}
		autoinfo.Delete(mu.Id)
	}
}

func (t *MediaUser) runner() {
	// t.mu.Lock()
	// defer t.mu.Unlock()

	// fmt.Println(t.Id, "------")
	// // 如果存在自动下载
	// if t.AutoDownload == 1 {

	// }

	// t.done = make(chan bool)
	// bs, err := t.StartGetter(
	// 	func(str string, bs *bs.Browser) {
	// 		dt := gjson.Parse(str)
	// 		if fans := dt.Get("fans").Int(); fans > 0 {
	// 			t.Fans = fans
	// 		}
	// 		if works := dt.Get("works").Int(); works > 0 {
	// 			t.Works = works
	// 		}
	// 		if local := dt.Get("local").String(); local != "" {
	// 			t.Local = local
	// 		}
	// 		if account := dt.Get("account").String(); account != "" {
	// 			t.Account = account
	// 		}
	// 		if t.AutoDownload == 1 && dt.Get("lists").Exists() {
	// 			var vids []string
	// 			for _, v := range dt.Get("lists").Array() {
	// 				vids = append(vids, v.String())
	// 			}
	// 			fmt.Println("找到带下载视频列表:", vids)
	// 			t.autodownload(vids)
	// 		}
	// 		t.LastDownTime = time.Now().Unix()
	// 		t.Save(nil)
	// 		t.Commpare()
	// 		bs.Close()
	// 		close(t.done)
	// 	},
	// 	func() {
	// 		eventbus.Bus.Publish("media_user_info", t)
	// 	},
	// 	func(url string) {

	// 	},
	// )
	// if err == nil {
	// 	t.Isruner = true
	// 	go func() {
	// 		sleep := 30 * time.Second
	// 		timer := time.NewTimer(sleep)
	// 		defer timer.Stop()
	// 		for {
	// 			select {
	// 			case <-timer.C:
	// 				if bs.IsArrive() {
	// 					bs.Close()
	// 					return
	// 				}
	// 				timer.Reset(sleep)
	// 			case <-t.done:
	// 				return
	// 			}
	// 		}
	// 	}()
	// }
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

	brows, _ := bs.BsManager.New(0, &bs.Options{
		Url:      runurl,
		JsStr:    runjs,
		Headless: false,
		Timeout:  time.Duration(time.Second * 30),
	}, true)

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
	dbs.DB().Model(&Media{}).Select("video_id").Where("user_id = ?", t.Id).Order("id DESC").First(&lastGetVideoId)

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
