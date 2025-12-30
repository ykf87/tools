package browser

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/chromedp/chromedp"
)

type BrowserClient interface {
	GetId() int64
	Open() error
	Close() error
	GetClient() *User
	// WorkDir() string
}

type AutoRun struct {
	Bs BrowserClient
	mu sync.Mutex

	allocCtx    context.Context
	allocCancel context.CancelFunc

	rootCtx    context.Context
	rootCancel context.CancelFunc

	opts []func(*chromedp.ExecAllocator)
}

func Client(binfile string, bs BrowserClient) *AutoRun {
	c := &AutoRun{Bs: bs}
	args := chromedp.DefaultExecAllocatorOptions[:]
	args = append(args, chromedp.UserDataDir(fmt.Sprintf("%s/%d", bs.GetClient().WorkDir(), bs.GetId())))
	args = append(args, chromedp.ExecPath(binfile), chromedp.Flag("enable-automation", false))
	c.opts = args
	return c
}

// 执行js
func (this *AutoRun) RunJs() {
	if err := this.Bs.Open(); err != nil {
		fmt.Println(err, "----")
		return
	}
	fmt.Println("执行脚本----")

	// this.allocCtx, this.allocCancel = chromedp.NewExecAllocator(context.Background(), this.opts...)
	// this.rootCtx, this.rootCancel = chromedp.NewContext(this.allocCtx)

	// if err := chromedp.Run(this.rootCtx, chromedp.Navigate("https://www.google.com")); err != nil {
	// 	this.allocCancel()
	// 	this.rootCancel()
	// 	fmt.Println("-----", err)
	// 	return
	// }
	// chromedp.Run(this.rootCtx,
	// 	chromedp.Navigate("https://www.google.com"))
	allocCtx, cancel := chromedp.NewRemoteAllocator(
		context.Background(),
		fmt.Sprintf("ws://127.0.0.1:%d", this.Bs.GetClient().ListenPort),
	)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	err := chromedp.Run(
		ctx,
		chromedp.Navigate("https://www.google.com"),
	)
	if err != nil {
		log.Println(err)
	}
}
