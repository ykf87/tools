package browser

import (
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

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

		case *runtime.EventConsoleAPICalled:
			if cb, ok := b.onConsole.Load().(func([]*runtime.RemoteObject)); ok {
				go safeCallConsole(cb, e.Args)
			}
		}
	})
}

func safeCallURL(cb func(string), url string) {
	defer func() { _ = recover() }()
	cb(url)
}

func safeCallConsole(cb func([]*runtime.RemoteObject), args []*runtime.RemoteObject) {
	defer func() { _ = recover() }()
	cb(args)
}

// func (b *Browser) OnConsole(cb func([]*runtime.RemoteObject)) {
// 	chromedp.ListenTarget(b.ctx, func(ev any) {
// 		if e, ok := ev.(*runtime.EventConsoleAPICalled); ok {
// 			cb(e.Args)
// 		}
// 	})
// }
