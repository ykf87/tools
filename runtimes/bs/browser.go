package bs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/i18n"

	"github.com/chromedp/chromedp"
	jsoniter "github.com/json-iterator/go"
)

var BASEPATH = config.BROWSERCACHE
var Json = jsoniter.ConfigCompatibleWithStandardLibrary

// 打开浏览器
func (b *Browser) OpenBrowser() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, err := MakeBrowserConfig(b.id, b.Opts.Language, b.Opts.Timezone, b.Opts.Proxy); err != nil {
		return err
	}
	allocOpts := make([]chromedp.ExecAllocatorOption, 0, len(chromedp.DefaultExecAllocatorOptions)+8)
	allocOpts = append(allocOpts, chromedp.DefaultExecAllocatorOptions[:]...)
	allocOpts = append(allocOpts,
		chromedp.ExecPath(b.Opts.ExecPath),
		chromedp.UserDataDir(b.Opts.UserDir),
		chromedp.WindowSize(b.Opts.Width, b.Opts.Height),
		chromedp.Flag("headless", b.Opts.Headless),
		chromedp.Flag("disable-gpu", b.Opts.Headless),
		chromedp.Flag("worker-id", fmt.Sprintf("%d", b.id)),
	)

	if b.Opts.Proxy != "" {
		allocOpts = append(allocOpts, chromedp.ProxyServer(b.Opts.Proxy))
	}
	if b.Opts.UserAgent != "" {
		allocOpts = append(allocOpts, chromedp.UserAgent(b.Opts.UserAgent))
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...any) {}))

	b.alloc = allocCancel
	b.ctx = ctx
	b.cancel = cancel

	// 下面代码是注入浏览器监控的
	// b.OnURLChange(func(url string) {
	// 	if b.JsStr != "" {
	// 		b.RunJs(b.JsStr)
	// 	}
	// })
	// b.OnConsole(func(args []*rt.RemoteObject) {
	// 	fmt.Println(args, "args-----")
	// })

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		allocCancel()
		return err
	}
	b.survival.Store(true)

	url := "about:blank"
	if b.Opts.Url != "" {
		url = b.Opts.Url
	}
	b.GoToUrl(url)
	// if b.opts.JsStr != "" {
	// 	b.RunJs(b.opts.JsStr)
	// }

	b.watchClose()
	go b.startEventLoop()

	go func(b *Browser) {
		<-ctx.Done()
		b.Close()
	}(b)
	return nil
}

func (b *Browser) GoToUrl(url string) error {
	return chromedp.Run(
		b.ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)
}

func (b *Browser) Run(actions ...chromedp.Action) error {
	if b.closed.Load() {
		return errors.New("browser closed")
	}

	ctx := b.ctx
	if b.Opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.Opts.Timeout)
		defer cancel()
	}

	return chromedp.Run(ctx, actions...)
}

func (b *Browser) RunJs(js string) (any, error) {
	if js == "" {
		js = b.Opts.JsStr
	} else {
		b.Opts.JsStr = js
	}
	var rs any
	if js != "" {
		if err := b.Run(chromedp.Evaluate(js, &rs)); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("js is empty")
	}

	return rs, nil
}

// 从chromedp上传
func (this *Browser) Upload(fls []string, clickNode, fileNode string) error {
	if clickNode == "" || fileNode == "" {
		return errors.New(i18n.T("上传数据不足"))
	}

	for i, v := range fls {
		v = strings.ReplaceAll(v, "\\", "/")
		if strings.Contains(v, config.MEDIAROOT) == false {
			v = filepath.Join(config.MEDIAROOT, v)
		}
		if _, err := os.Stat(v); err != nil {
			return err
		}
		fls[i] = v
	}

	err := this.Run(
		chromedp.Click(clickNode),
		chromedp.Sleep(1*time.Second),
		chromedp.SetUploadFiles(fileNode, fls),
	)
	return err
}

// 输入内容
func (this *Browser) InputTxt(text, clickNode string) error {
	if text == "" || clickNode == "" {
		return errors.New(i18n.T("输入数据不足"))
	}

	var backs []chromedp.Action
	for i := len(text); i > 0; i-- {
		backs = append(backs, chromedp.SendKeys(clickNode, "\b"))
	}

	this.Run(
		chromedp.Click(clickNode),
		chromedp.Sleep(time.Second*2),
	)
	this.Run(backs...)
	return this.Run(
		chromedp.Sleep(time.Second*1),
		chromedp.SendKeys(clickNode, text, chromedp.NodeVisible),
	)
}

// 当前浏览器是否存活
func (this *Browser) IsArrive() bool {
	return this.survival.Load() && !this.closed.Load()
}
