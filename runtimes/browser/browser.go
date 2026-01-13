// 用法
// mgr := browser.NewManager(this.Bs.WorkDir()) 传入缓存根目录
//
//	b, _ := mgr.New(fmt.Sprintf("%d", this.Id), browser.Options{
//		ExecPath: browser.BROWSERFILE,// 执行文件地址
//		Headless: false,// 是否有窗口运行, false为打开窗口
//		Width:    this.Width,
//		Height:   this.Height,
//		Temp:     true,
//		Timeout:  30 * time.Second,
//		Proxy:    proxyUrl,
//	})
//
//	b.OnURLChange(func(url string) {// 监听浏览器地址改变
//		fmt.Println("URL:", url)
//	})
//
//	go func() {// 监听浏览器被手动关闭
//		<-b.OnClosed()
//		eventbus.Bus.Publish("browser-close", this.Bs)
//	}()
//
// b.Run(// 启动
//
//	chromedp.Navigate("https://www.browserscan.net/"),
//
// )
package browser

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/i18n"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Browser struct {
	id     string
	opts   Options
	ctx    context.Context
	cancel context.CancelFunc
	alloc  context.CancelFunc
	once   sync.Once
	closed atomic.Bool

	onClosed    chan struct{} // ✅ 新增
	onURLChange atomic.Value  // func(string)
	onConsole   atomic.Value  // func([]*runtime.RemoteObject)
	JsStr       string        `json:"js_str" gorm:"-" form:"-"` // 执行的js
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

func (b *Browser) Close() {
	b.once.Do(func() {
		b.closed.Store(true)

		b.cancel()
		b.alloc()

		close(b.onClosed)

		if b.opts.Temp {
			_ = os.RemoveAll(b.opts.UserDir)
		}
	})
}

func (b *Browser) IsClosed() bool {
	return b.closed.Load()
}

func (b *Browser) watchClose() {
	go func() {
		<-b.ctx.Done()

		// 如果已经主动 Close，就不再重复处理
		if b.closed.Load() {
			return
		}

		b.once.Do(func() {
			b.closed.Store(true)
			close(b.onClosed)

			if b.opts.Temp {
				_ = os.RemoveAll(b.opts.UserDir)
			}
		})
	}()
}

func (b *Browser) OnClosed() <-chan struct{} {
	return b.onClosed
}

func (b *Browser) OnURLChange(cb func(string)) {
	b.onURLChange.Store(cb)
}

func (b *Browser) OnConsole(cb func([]*runtime.RemoteObject)) {
	b.onConsole.Store(cb)
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
