package medias

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/proxys"
	"tools/runtimes/db/task"
	"tools/runtimes/funcs"
	"tools/runtimes/proxy"
	"tools/runtimes/runner"
	"tools/runtimes/videos/downloader/parser"

	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

type Options struct {
	MUID        int64
	Width       int
	Height      int
	ClientType  int
	ClientID    int64
	IsShow      bool
	Url         string
	Js          string
	Proxy       string
	UA          string
	Lang        string
	Timezone    string
	Timeout     time.Duration
	mu          *MediaUser
	ProxyConfig *proxy.ProxyConfig
	// runner      *tasklog.TaskRunner
	ctx context.Context
	tr  *task.TaskRun
}

var MURUNNERS sync.Map

func GetOptions(mu *MediaUser, ctx context.Context, tr *task.TaskRun) (*Options, error) {
	// if opt, err := getRunnerOption(mu.Id); err == nil {
	// 	return opt, nil
	// }
	opt := &Options{
		MUID:    mu.Id,
		mu:      mu,
		Timeout: time.Second * 300,
		ctx:     ctx,
		tr:      tr,
		IsShow:  false,
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
func (opt *Options) Start(downloads bool) error {
	var opts any

	switch opt.ClientType {
	case 1:
		opts = runner.GenPhoneOpt()
		// return fmt.Errorf("手机端暂未支持")
	case 2:
		opts = runner.GenHttpOpt()
	// return fmt.Errorf("HTTP端暂未支持")
	default:
		opts = runner.GenWebOpt(
			opt.ctx,
			opt.ClientID,
			opt.IsShow,
			opt.Url,
			opt.Js,
			opt.ProxyConfig,
			opt.Timeout,
			opt.Width,
			opt.Height,
			opt.Lang,
			opt.Timezone,
		)
	}

	r, err := runner.GetRunner(opt.ClientType, opts)
	if err != nil {
		return err
	}

	err = r.Start(opt.Timeout, func(str string) error {
		r.Stop()
		opt.tr.SentMsg("信息获取成功,正在解析...", 0, false)
		return opt.ParseInfos(str, downloads)
	})
	return err
}

// 将js获取到的内容格式化到用户
func (opt *Options) ParseInfos(str string, downloads bool) error {
	gs := gjson.Parse(str)
	if downloads == true {
		if !gs.Get("lists").Exists() || len(gs.Get("lists").Array()) < 1 {
			opt.tr.SentMsg("找不到下载列表", 1, true)
			return nil
		}

		var vids []string
		mmp := make(map[string]string)
		for _, v := range gs.Get("lists").Array() {
			sr := v.Get("url").String()
			vtem := strings.Split(sr, "/")
			vid := vtem[len(vtem)-1]
			vids = append(vids, vid)
			mmp[vid] = sr
		}

		var inMedia []string
		dbs.DB().Model(&Media{}).
			Select("video_id").
			Where("user_id = ? and platform = ? and video_id in ?", opt.MUID, opt.mu.Platform, vids).
			Find(&inMedia)
		if opt.Proxy == "" && opt.ProxyConfig != nil {
			opt.Proxy = opt.ProxyConfig.Listened()
		}
		var transport *http.Transport
		if opt.Proxy != "" {
			if proxyURL, err := url.Parse(opt.Proxy); err == nil {
				transport = &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
			}
		}

		if len(vids) > 0 {
			var idx float64
			urlstotal := float64(len(vids))
			opt.tr.ReportSchedule(urlstotal, 0)
			opt.tr.SentMsg("开始下载视频...", 0, false)

			for vid, url := range mmp {
				idx = idx + 1
				if slices.Contains(inMedia, vid) {
					opt.tr.ReportSchedule(urlstotal, idx)
					continue
				}
				select {
				case <-opt.ctx.Done():
					return nil
				default:
					if !strings.HasPrefix(url, "http") {
						url = fmt.Sprintf("https:%s", url)
					}
					parseRes, err := parser.ParseVideoShareUrlByRegexp(url, transport)
					if err != nil {
						// fmt.Println("解析地址错误:", err)
						opt.tr.SentMsg("解析地址错误:"+err.Error(), 1, true)
						opt.tr.ReportSchedule(urlstotal, idx)
					} else {
						if parseRes.VideoUrl != "" {
							vtem := strings.Split(url, "/")
							mp, err := MKDBNameID( // 构建文件存储目录
								"autodownload/"+opt.mu.Name,
								0,
							)
							if err != nil {
								return err
							}
							m := &Media{
								Platform: opt.mu.Platform,
								UserId:   opt.MUID,
								VideoID:  vtem[len(vtem)-1],
								Title:    parseRes.Title,
								Url:      opt.Url,
								UrlMd5:   funcs.Md5String(opt.Url),
								Path:     "autodownload/" + opt.mu.Name,
								PathID:   mp.ID,
							}
							m.DownMediaFiles([]string{parseRes.VideoUrl}, opt.Proxy, nil)

							path := opt.mu.DefDirName("") //fmt.Sprintf(".auto/%s%d", opt.mu.Uuid, opt.mu.Id)
							md, err := DownLoadVideo(url, []string{parseRes.VideoUrl}, path, "", opt.Proxy, func(percent float64, downloaded, total int64) {})
							if err == nil {
								md.UserId = opt.MUID
								vtem := strings.Split(url, "/")
								md.VideoID = vtem[len(vtem)-1]
								md.Platform = opt.mu.Platform
								md.Title = parseRes.Title
								dbs.Write(func(tx *gorm.DB) error {
									// return md.Save(md, tx)
									return nil
								})
								opt.tr.ReportSchedule(urlstotal, idx)
								opt.tr.SentMsg("成功下载一个视频", 0, false)
							} else {
								opt.tr.ReportSchedule(urlstotal, idx)
								opt.tr.SentMsg("下载视频失败:"+err.Error(), 0, false)
							}
						} else {
							opt.tr.ReportSchedule(urlstotal, idx)
							opt.tr.SentMsg("找不到视频地址:", 0, false)
						}
					}
					time.Sleep(time.Duration(funcs.RandomNumber(3, 10)) * time.Second)
				}
			}
		} else {
			// fmt.Println("没有新的视频以供下载!")
			// return errors.New("没有新的视频以供下载!")
			return nil
		}
	}

	opt.mu.ParseUserInfoData(str)
	return nil
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
