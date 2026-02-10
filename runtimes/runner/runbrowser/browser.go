package runbrowser

import (
	"context"
	"errors"
	"fmt"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/mainsignal"
)

type RunBrowser struct {
	Browser *bs.Browser
	ctx     context.Context
	cancle  context.CancelFunc
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

func (r *RunBrowser) Start(timeout time.Duration, callback func(str string) error) error {
	if r.Browser.Opts.Ctx == nil {
		r.ctx, r.cancle = context.WithTimeout(mainsignal.MainCtx, timeout)
		r.Browser.Opts.Ctx = r.ctx
	} else {
		r.ctx, r.cancle = context.WithTimeout(r.Browser.Opts.Ctx, timeout)
	}

	if r.Browser.Opts.Proxy == "" && r.Browser.Opts.Pc != nil {
		if _, err := r.Browser.Opts.Pc.Run(false); err == nil {
			r.Browser.Opts.Proxy = r.Browser.Opts.Pc.Listened()
		}
	}

	if err := r.Browser.OpenBrowser(); err != nil {
		return fmt.Errorf("浏览器打开失败: %s", err.Error())
	}
	r.Browser.Opts.Msg = make(chan string)

	var idx int
	bbctx := r.Browser.GetCtx()
	var err error
	for {
		select {
		case <-r.ctx.Done():
			err = errors.New("超时")
			fmt.Println("超时结束")
			goto BREAK
		case msg := <-r.Browser.Opts.Msg:
			if callback != nil {
				if errr := callback(msg); errr != nil {
					idx++
					if idx >= 5 {
						err = errr
						goto BREAK
					}
					r.Start(timeout, callback)
				} else {
					r.Stop()
				}
			}
		case <-bbctx.Done():
			if errors.Is(bbctx.Err(), context.DeadlineExceeded) {
				err = errors.New("浏览器超时")
				fmt.Println("浏览器超时了")
			}
			if errors.Is(bbctx.Err(), context.Canceled) {
				err = errors.New("浏览器意外关闭")
				fmt.Println("浏览器被关闭")
			}
			goto BREAK
		}
	}
BREAK:
	r.cancle()
	return err
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
