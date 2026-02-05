package medias

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/proxys"
	"tools/runtimes/db/tasklog"
	"tools/runtimes/proxy"
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

func getOptions(mu *MediaUser) (*Options, error) {
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

func (opt *Options) stop() {
	if opt.runner != nil {
		opt.runner.Runner.Stop()
	}
	MURUNNERS.Delete(opt.MUID)
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
	cls := opt.mu.GetClients()
	var ct int
	var cid int64
	if clientID, clientType, ok := RandomInt64Fast(cls); ok {
		opt.ClientID, opt.ClientType = clientID, clientType
	}

	var proxyID int64
	var proxyCongig string
	if cid > 0 {
		switch ct {
		case 0:
			bbs, err := browserdb.GetBrowserById(cid)
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

	// dt := gjson.Parse(msg)
	// if fans := dt.Get("fans").Int(); fans > 0 {
	// 	mu.Fans = fans
	// }
	// if works := dt.Get("works").Int(); works > 0 {
	// 	mu.Works = works
	// }
	// if local := dt.Get("local").String(); local != "" {
	// 	mu.Local = local
	// }
	// if account := dt.Get("account").String(); account != "" {
	// 	mu.Account = account
	// }
	// if dt.Get("lists").Exists() {
	// 	var vids []string
	// 	for _, v := range dt.Get("lists").Array() {
	// 		vids = append(vids, v.String())
	// 	}
	// 	fmt.Println("找到带下载视频列表:", vids)
	// 	mu.autodownload(vids)
	// }
	if tr.Msg != nil {
		tr.Msg <- "本次任务完成-来自getter"
	}
	return nil
}

// 自动下载用户视频
func (opt *Options) FmtDownload(msg string, tr *tasklog.TaskRunner) error {
	tr.SetMsg("信息获取成功,正在处理获得的信息...")
	fmt.Println(msg, "-----")
	time.Sleep(time.Second * 15)
	tr.Bss.Close()
	return nil
}

// 启动下载任务
func (opt *Options) Start() error {
	switch opt.ClientType {
	case 0:
		if err := opt.runner.StartWeb(&bs.Options{
			Url:      opt.Url,
			JsStr:    opt.Js,
			Headless: opt.Headless,
			Timeout:  opt.Timeout,
			Pc:       opt.ProxyConfig,
			ID:       opt.ClientID,
		}, opt.Timeout); err != nil {
			return err
		}
		MURUNNERS.Store(opt.MUID, opt)
		opt.runner.SetMsg("任务加入队列准备执行...")
		opt.runner.Runner.Every(time.Duration(opt.mu.DownFreq) * time.Minute).SetError(func(err error) {
			fmt.Println("执行失败:", err)
		}).SetMaxTry(5).RunNow()
	case 1:
		return fmt.Errorf("手机端暂未支持")
	case 2:
		return fmt.Errorf("HTTP端暂未支持")
	}
	return nil
}
