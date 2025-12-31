package browser

import (
	"fmt"
	"time"
	"tools/runtimes/eventbus"

	"github.com/chromedp/chromedp"
)

// 打开vb浏览器
// workerDir 	- 浏览器缓存路径
// workerId		- 唯一标识
// width 		- 屏幕宽度
// height 		- 屏幕高度
// proxy		- 代理地址
// openedFun	- 成功开启后的回调
// closeFun 	- 关闭后的回调
func OpenBrowser(workerDir string, workerId string, width, height int, proxy string, openedFun func(bs *Browser), closeFun func(workid string)) error {
	mgr := NewManager(workerDir)
	b, _ := mgr.New(workerId, Options{
		ExecPath: BROWSERFILE,
		Headless: false,
		Width:    width,
		Height:   height,
		Temp:     true,
		Timeout:  30 * time.Second,
		Proxy:    proxy,
	})
	b.OnURLChange(func(url string) { // 当url地址改变后
		fmt.Println("URL:", url)
	})
	go func() {
		<-b.OnClosed()
		if closeFun != nil {
			closeFun(workerId)
		}
		eventbus.Bus.Publish("browser-close", workerId)
	}()
	if err := b.Run(
		chromedp.Navigate("about:blank"),
	); err != nil {
		return err
	}

	Running.Store(workerId, b)
	if openedFun != nil {
		openedFun(b)
	}
	return nil
}
