package runweb

import (
	"context"
	"fmt"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/scheduler"

	"github.com/chromedp/cdproto/runtime"
	"github.com/tidwall/gjson"
)

type Option struct {
	Url      string
	ID       int64
	Js       string
	Headless bool
	Timeout  time.Duration
	Ctx      context.Context
	Callback func(string) error
	OnError  func(error, *bs.Browser)
	OnClose  func()
	OnChange func(string, *bs.Browser) error
}

type runweb struct {
	callback  func()
	scheduler *scheduler.Runner
	opt       *Option
	bs        *bs.Browser
	chs       chan struct{}
}

func New(callback func(), opt *Option) *runweb {
	return &runweb{
		callback: callback,
		opt:      opt,
	}
}

func (t *runweb) SetRunner(s *scheduler.Runner) {
	t.scheduler = s
}

func (t *runweb) Start(ctx context.Context) error {
	bs, err := bs.BsManager.New(t.opt.ID, bs.Options{
		Url:      t.opt.Url,
		JsStr:    t.opt.Js,
		Headless: t.opt.Headless,
		Timeout:  t.opt.Timeout,
		Ctx:      t.opt.Ctx,
	}, true)
	if err != nil {
		return err
	}
	t.chs = make(chan struct{})
	fmt.Println("执行任务:", t.scheduler.GetID())
	t.bs = bs

	t.bs.OnClosed(t.OnClose)
	t.bs.OnURLChange(t.OnChange)

	t.bs.OnConsole(func(args []*runtime.RemoteObject) {
		for _, arg := range args {
			if arg.Value != nil {
				gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
				if gs.Get("type").String() == "kaka" {
					if t.opt.Callback != nil {
						if err := t.opt.Callback(gs.Get("data").String()); err != nil {
							t.OnError(err)
						}
					}
				}
			}
		}
	})

	t.bs.OpenBrowser()

	if t.opt.Url != "" {
		t.bs.GoToUrl(t.opt.Url)
	}
	if t.opt.Js != "" {
		t.bs.RunJs(t.opt.Js)
	}
	<-t.chs
	return nil
}

func (t *runweb) OnError(err error) {
	if t.opt.OnError != nil {
		t.opt.OnError(err, t.bs)
		return
	}
	t.bs.Close()
}
func (t *runweb) OnClose() {
	defer t.callback()
	if t.opt.OnClose != nil {
		t.opt.OnClose()
	}
	t.chs <- struct{}{}
	fmt.Println("任务结束,总执行时间:", t.scheduler.GetTotalTime(), ".重试:", t.scheduler.GetTryTimers())
}
func (t *runweb) OnChange(str string) {
	if t.opt.OnChange != nil {
		if err := t.opt.OnChange(str, t.bs); err != nil {
			t.bs.Close()
		}
	}
}
