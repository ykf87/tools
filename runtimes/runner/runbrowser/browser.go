package runbrowser

import (
	"errors"
	"fmt"
	"tools/runtimes/bs"
	"tools/runtimes/mainsignal"

	"github.com/chromedp/cdproto/runtime"

	"github.com/tidwall/gjson"
)

type RunBrowser struct {
	Browser *bs.Browser
}

func New(opt any, wait bool) (*RunBrowser, error) {
	myopt, ok := opt.(*bs.Options)
	if !ok {
		return nil, errors.New("浏览器opt配置错误")
	}
	bbs, err := bs.BsManager.New(myopt.ID, myopt, wait)
	if err != nil {
		return nil, err
	}

	return &RunBrowser{
		Browser: bbs,
	}, nil
}

func (r *RunBrowser) Start() error {
	if r.Browser.Opts.Ctx == nil {
		r.Browser.Opts.Ctx = mainsignal.MainCtx
	}

	if r.Browser.Opts.Proxy == "" && r.Browser.Opts.Pc != nil {
		if _, err := r.Browser.Opts.Pc.Run(false); err == nil {
			r.Browser.Opts.Proxy = r.Browser.Opts.Pc.Listened()
		}
	}

	r.Browser.OnConsole(func(args []*runtime.RemoteObject) {
		for _, arg := range args {
			if arg.Value != nil {
				gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
				if gs.Get("version").String() == "" {
					continue
				}
				switch gs.Get("type").String() {
				case "upload": // 调用系统的上传功能
					var fls []string
					for _, v := range gs.Get("data.files").Array() {
						fls = append(fls, v.String())
					}
					r.Browser.Upload(fls, gs.Get("data.node").String(), gs.Get("data.upnode").String())
					r.sendMsg("上传文件")
				case "input": // 输入
					r.Browser.InputTxt(gs.Get("data.text").String(), gs.Get("data.node").String())
					r.sendMsg("输入数据")
				default:
					r.sendMsg(gs.Get("data").String())
				}
			}
		}
	})

	if err := r.Browser.OpenBrowser(); err != nil {
		return fmt.Errorf("浏览器打开失败: %s", err.Error())
	}

	return nil
}

func (r *RunBrowser) sendMsg(str string) {
	if r.Browser.Opts.Msg != nil {
		select {
		case r.Browser.Opts.Msg <- str:
		case <-r.Browser.GetCtx().Done():
		}
	}
}

func (r *RunBrowser) Stop() {
	r.Browser.Close()
}

func IsRuning(id int64) bool {
	return bs.BsManager.IsArride(id)
}
