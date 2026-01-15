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

	if _, err := MakeBrowserConfig(b.id, b.opts.Language, b.opts.Timezone, b.opts.Proxy); err != nil {
		return err
	}
	allocOpts := make([]chromedp.ExecAllocatorOption, 0, len(chromedp.DefaultExecAllocatorOptions)+8)
	allocOpts = append(allocOpts, chromedp.DefaultExecAllocatorOptions[:]...)
	allocOpts = append(allocOpts,
		chromedp.ExecPath(b.opts.ExecPath),
		chromedp.UserDataDir(b.opts.UserDir),
		chromedp.WindowSize(b.opts.Width, b.opts.Height),
		chromedp.Flag("headless", b.opts.Headless),
		chromedp.Flag("disable-gpu", b.opts.Headless),
		chromedp.Flag("worker-id", fmt.Sprintf("%d", b.id)),
	)

	if b.opts.Proxy != "" {
		allocOpts = append(allocOpts, chromedp.ProxyServer(b.opts.Proxy))
	}
	if b.opts.UserAgent != "" {
		allocOpts = append(allocOpts, chromedp.UserAgent(b.opts.UserAgent))
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
	if b.opts.Url != "" {
		url = b.opts.Url
	}
	go b.GoToUrl(url)

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
	)
}

func (b *Browser) Run(actions ...chromedp.Action) error {
	if b.closed.Load() {
		return errors.New("browser closed")
	}

	ctx := b.ctx
	if b.opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.opts.Timeout)
		defer cancel()
	}

	return chromedp.Run(ctx, actions...)
}

func (b *Browser) RunJs(js string) error {
	if js == "" {
		js = b.JsStr
	} else {
		b.JsStr = js
	}
	var rs string
	return b.Run(chromedp.Evaluate(js, &rs))
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
