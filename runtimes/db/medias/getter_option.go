package medias

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/proxys"
	"tools/runtimes/db/tasklog"
	"tools/runtimes/funcs"
	"tools/runtimes/listens/ws"
	"tools/runtimes/proxy"
	"tools/runtimes/videos/downloader/parser"

	"github.com/tidwall/gjson"
)

type Options struct {
	MUID        int64
	Width       int
	Height      int
	ClientType  int
	ClientID    int64
	Headless    bool
	Url         string
	Js          string
	Proxy       string
	UA          string
	Lang        string
	Timezone    string
	Timeout     time.Duration
	mu          *MediaUser
	ProxyConfig *proxy.ProxyConfig
	runner      *tasklog.TaskRunner
}

var MURUNNERS sync.Map

func GetOptions(mu *MediaUser) (*Options, error) {
	if opt, err := getRunnerOption(mu.Id); err == nil {
		return opt, nil
	}
	opt := &Options{
		MUID:    mu.Id,
		mu:      mu,
		Timeout: time.Second * 300,
	}
	if err := opt.getJsAndUrl(); err != nil {
		return nil, err
	}

	if err := opt.getProxyAndClient(); err != nil {
		return nil, err
	}
	return opt, nil
}

// 启动周期性自动任务
func (opt *Options) Start() error {
	switch opt.ClientType {
	case 0:
		// runner.GenWebOpt()
		bsopt := &bs.Options{
			Url:      opt.Url,
			JsStr:    opt.Js,
			Headless: opt.Headless,
			Timeout:  opt.Timeout,
			Pc:       opt.ProxyConfig,
			ID:       opt.ClientID,
		}
		if opt.runner != nil { // 周期性任务
			if err := opt.runner.StartWeb(bsopt, opt.Timeout); err != nil {
				return err
			}
			MURUNNERS.Store(opt.MUID, opt)
			opt.runner.SetMsg("任务加入队列准备执行...")
			opt.runner.Runner.Every(time.Duration(opt.mu.DownFreq) * time.Minute).SetError(func(err error) {
				fmt.Println("执行失败:", err)
			}).SetMaxTry(5).Run()
		} else {
			// opt.runner = TaskLogger.Append(
			// 	mainsignal.MainCtx,
			// 	fmt.Sprintf("muautodown-%d", mu.Id),
			// 	fmt.Sprintf("%s 自动下载", mu.Name),
			// 	opt.FmtDownload,
			// )
		}
	case 1:
		return fmt.Errorf("手机端暂未支持")
	case 2:
		return fmt.Errorf("HTTP端暂未支持")
	}
	return nil
}

// 启动每天的定点任务
func (opt *Options) StartDailyRandomAt(h, m, s, j int) error {
	switch opt.ClientType {
	case 0:
		bsopt := &bs.Options{
			Url:      opt.Url,
			JsStr:    opt.Js,
			Headless: opt.Headless,
			Timeout:  opt.Timeout,
			Pc:       opt.ProxyConfig,
			ID:       opt.ClientID,
		}
		if opt.runner != nil { // 定时任务
			if err := opt.runner.StartWeb(bsopt, opt.Timeout); err != nil {
				return err
			}
			MURUNNERS.Store(opt.MUID, opt)
			opt.runner.SetMsg("任务加入定点队列准备执行...")
			opt.runner.Runner.DailyRandomAt(h, m, s, j, nil).SetError(func(err error) {
				fmt.Println("执行失败:", err)
			}).SetMaxTry(5).Run()
		}
	case 1:
		return fmt.Errorf("手机端暂未支持")
	case 2:
		return fmt.Errorf("HTTP端暂未支持")
	}
	return nil
}

// 停止任务
func (opt *Options) Stop() {
	if opt.runner != nil {
		opt.runner.Runner.Stop()
	}
	MURUNNERS.Delete(opt.MUID)
}

func RandomInt64Fast(m map[int][]int64) (int64, int, bool) {
	if len(m) == 0 {
		return 0, 0, false
	}

	keys := make([]int, 0, len(m))
	for k, v := range m {
		if len(v) > 0 {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return 0, 0, false
	}
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))

	k := keys[r.Intn(len(keys))]
	arr := m[k]

	return arr[r.Intn(len(arr))], k, true
}

func getRunnerOption(id int64) (*Options, error) {
	if opto, ok := MURUNNERS.Load(id); ok {
		if opt, ok := opto.(*Options); ok {
			return opt, nil
		}
	}
	return nil, fmt.Errorf("未执行")
}

func (opt *Options) getJsAndUrl() error {
	var jscode string
	var runurl string
	switch opt.mu.Platform {
	case "douyin":
		jscode = "douyin-info"
		runurl = fmt.Sprintf("https://www.douyin.com/user/%s", opt.mu.Uuid)
	default:
		err := fmt.Errorf("暂时不支持 %s 获取获取信息", opt.mu.Platform)
		return err
	}

	js := jses.GetJsByCode(jscode)
	if js == nil || js.ID < 1 {
		err := fmt.Errorf("获取账号信息脚本不存在")
		return err
	}
	opt.Url = runurl
	opt.Js = js.GetContent(nil)
	return nil
}

// 获取用户使用的设备和代理节点
func (opt *Options) getProxyAndClient() error {
	opt.ClientType, opt.ClientID = opt.mu.GetCanUseClient()

	var proxyID int64
	var proxyCongig string
	if opt.ClientID > 0 {
		switch opt.ClientType {
		case 0:
			bbs, err := browserdb.GetBrowserById(opt.ClientID)
			if err != nil {
				return err
			}
			proxyID = bbs.Proxy
			proxyCongig = bbs.ProxyConfig
			opt.Width = bbs.Width
			opt.Height = bbs.Height
			opt.Lang = bbs.Lang
			opt.Timezone = bbs.Timezone
		case 1:
		case 2:
		}
	} else {
		if ps := opt.mu.GetProxys(); len(ps) > 0 {
			proxyID = ps[rand.Intn(len(ps))]
		}
	}
	var err error
	if proxyID > 0 {
		if opt.ProxyConfig, err = proxys.GetProxyConfigByID(proxyID); err != nil {
			return err
		}
	} else if proxyCongig != "" {
		if opt.ProxyConfig, err = proxy.Client(proxyCongig, "", 0, ""); err != nil {
			return err
		}
	}

	return nil
}

// 自动更新用户信息
func (opt *Options) FmtInfo(msg string, tr *tasklog.TaskRunner) error {
	tr.Msg <- "开始执行任务"
	// fmt.Println(msg, "------ run")

	if tr.Msg != nil {
		tr.Msg <- "本次任务完成-来自getter"
	}
	return nil
}

// 自动下载用户视频
func (opt *Options) FmtDownload(msg string, tr *tasklog.TaskRunner) error {
	tr.SetMsg("信息获取成功,正在处理获得的信息...")

	dt := gjson.Parse(msg)
	if fans := dt.Get("fans").Int(); fans > 0 {
		opt.mu.Fans = fans
	}
	if works := dt.Get("works").Int(); works > 0 {
		opt.mu.Works = works
	}
	if local := dt.Get("local").String(); local != "" {
		opt.mu.Local = local
	}
	if account := dt.Get("account").String(); account != "" {
		opt.mu.Account = account
	}
	if dt.Get("lists").Exists() {
		var vids []string
		for _, v := range dt.Get("lists").Array() {
			vids = append(vids, v.String())
		}
		opt.autodownload(vids)
	}

	tr.Bss.Close()
	return nil
}

// 下载视频
func (opt *Options) autodownload(srcs []string) {
	if len(srcs) < 1 {
		opt.runner.SetErrMsg("账号主页没有视频.")
		return
	}
	var lastVid string
	dbs.Model(&Media{}).Select("video_id").
		Where("user_id = ?", opt.mu.Id).
		Where("video_id is not null or video_id != ''").
		Order("id DESC").First(&lastVid)

	var getterSrcs []string
	if lastVid != "" {
		for _, v := range srcs {
			if v == "" || !strings.Contains(v, "/") {
				continue
			}
			ssr := strings.Split(v, "/")
			vid := ssr[len(ssr)-1]
			if vid == lastVid {
				break
			}
			getterSrcs = append(getterSrcs, v)
		}
	} else {
		defLen := 5
		opt.runner.SetMsg(fmt.Sprintf("该账号未下载过视频,默认下载最新的 %d 条", defLen))
		brkLen := defLen - 1
		for k, v := range srcs {
			getterSrcs = append(getterSrcs, v)
			if k >= brkLen {
				break
			}
		}
	}

	if len(getterSrcs) > 0 {
		wg := new(sync.WaitGroup)
		var transport *http.Transport
		if opt.Proxy != "" {
			if proxyURL, err := url.Parse(opt.Proxy); err == nil {
				transport = &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
			}
		}
		for _, v := range getterSrcs {
			wg.Go(func() {
				parseRes, err := parser.ParseVideoShareUrlByRegexp(v, transport)
				if err != nil {
					opt.runner.SentErr("解析视频地址错误: " + err.Error() + " - " + v)
					return
				}
				if parseRes.VideoUrl != "" {
					fn := funcs.Md5String(parseRes.VideoUrl)
					path := fmt.Sprintf(".auto/%s%d", opt.mu.Uuid, opt.mu.Id)
					md, err := DownLoadVideo(parseRes.VideoUrl, path, "", opt.Proxy, func(percent float64, downloaded, total int64) {
						fmt.Printf("\r下载进度: %.2f%%", percent)
						opt.runner.Total = float64(total)
						opt.runner.Doned = float64(downloaded)
						opt.runner.Sent(fmt.Sprintf("下载进度: %.2f%%", percent))

						dbk := new(pms)
						dbk.DownFile = fn
						dbk.Fmt = fmt.Sprintf("%.2f%%", percent)
						dbk.Num = percent
						dbk.Dir = false
						dbk.Cover = parseRes.CoverUrl
						dbk.Name = fn
						dbk.Platform = parseRes.Platform
						ws.SentBus(opt.mu.AdminID, "video-download", dbk, "")
					})

					if err != nil {
						opt.runner.SetErrMsg("视频下载失败: " + err.Error())
					}

					md.Title = parseRes.Title
					md.Platform = parseRes.Platform
					md.VideoID = parseRes.VideoID
					md.UserId = opt.mu.Id

					if err := md.Save(nil); err != nil {
						opt.runner.SetErrMsg("视频落库失败: " + err.Error())
						return
					}
				}
			})
		}
		wg.Wait()
		opt.mu.LastDownTime = time.Now().Unix()
	}

	// opt.mu.Id
}
