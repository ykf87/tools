package bs

import (
	"os"
	"sync"

	"github.com/chromedp/cdproto/page"
	rt "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

var OpendBrowser sync.Map

func (b *Browser) watchClose() {
	go func() {
		<-b.ctx.Done()
		b.Close()
	}()
}

func (b *Browser) startEventLoop() {
	chromedp.ListenTarget(b.ctx, func(ev any) {
		if b.closed.Load() {
			return
		}

		switch e := ev.(type) {

		case *page.EventFrameNavigated:
			if e.Frame.ParentID == "" {
				if cb, ok := b.onURLChange.Load().(func(string)); ok {
					go safeCallURL(cb, e.Frame.URL)
				}
			}

		case *rt.EventConsoleAPICalled:
			if cb, ok := b.onConsole.Load().(func([]*rt.RemoteObject)); ok {
				go safeCallConsole(cb, e.Args)
			}
		}
	})
}

func safeCallURL(cb func(string), url string) {
	defer func() { _ = recover() }()
	cb(url)
}

func safeClosed(cb func()) {
	defer func() { _ = recover() }()
	cb()
}

func safeCallConsole(cb func([]*rt.RemoteObject), args []*rt.RemoteObject) {
	defer func() { _ = recover() }()
	cb(args)
}

func (b *Browser) Close() {
	b.mu.Lock()
	defer func() {
		if b.Locker != nil {
			b.Locker <- 1
		}
		b.mu.Unlock()
		<-maxNumsCh
	}()

	BsManager.mu.Lock()
	delete(BsManager.browsers, b.ID)
	BsManager.mu.Unlock()

	if _, ok := OpendBrowser.Load(b.ID); ok {
		OpendBrowser.Delete(b.ID)
	}

	if b.survival.Load() {
		b.cancel()
		b.alloc()

		if b.Opts.Temp {
			_ = os.RemoveAll(b.Opts.UserDir)
		}
		b.survival.Store(false)
		// 通知ws浏览器被关闭
		if cb, ok := b.onClose.Load().(func()); ok {
			go safeClosed(cb)
		}
		if b.Opts.Pc != nil {
			b.Opts.Pc.Close(false)
		}
		if b.Opts.Msg != nil {
			close(b.Opts.Msg)
			b.Opts.Msg = nil
		}
	}

	b.closed.Store(true)
}

// func (b *Browser) OnClosed() <-chan struct{} {
// 	return b.onClosed
// }

func (b *Browser) OnURLChange(cb func(string)) {
	b.onURLChange.Store(cb)
}

func (b *Browser) OnClosed(cb func()) {
	b.onClose.Store(cb)
}

func (b *Browser) OnConsole(cb func([]*rt.RemoteObject)) {
	b.onConsole.Store(cb)
}

func ConseleFun(callback func(str string)) {

}
