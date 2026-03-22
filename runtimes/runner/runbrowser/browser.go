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

func (r *RunBrowser) Start(
	timeout time.Duration,
	callback func(msg, data string) error,
	errcallback func(msg string),
	msgcallback func(msg string),
) error {
	if callback == nil {
		return errors.New("未设置回调函数")
	}
	r.Browser.Opts.Callback = callback
	if errcallback != nil {
		r.Browser.Opts.ErrCallback = errcallback
	}
	if msgcallback != nil {
		r.Browser.Opts.MsgCallback = msgcallback
	}
	if r.Browser.Opts.Ctx == nil {
		r.ctx, r.cancle = context.WithTimeout(mainsignal.MainCtx, timeout)
		r.Browser.Opts.Ctx = r.ctx
	} else {
		r.ctx, r.cancle = context.WithTimeout(r.Browser.Opts.Ctx, timeout)
	}

	if err := r.Browser.OpenBrowser(); err != nil {
		return fmt.Errorf("浏览器打开失败: %s", err.Error())
	}
	// r.Browser.Opts.Msg = make(chan string)
	// if r.Browser.Opts.Msg == nil {
	// 	r.Browser.Opts.Msg = make(chan string)
	// }

	// var idx int
	bbctx := r.Browser.GetCtx()

	defer func() {
		r.Stop()
		r.cancle()
	}()
	for {
		select {
		case <-r.ctx.Done():
			return nil
		// case msg := <-r.Browser.Opts.Msg:
		// 	r.Msg(msg)
		case <-bbctx.Done():
			if errors.Is(bbctx.Err(), context.DeadlineExceeded) {
				return errors.New("浏览器超时")
			}
		}
	}
}

func (r *RunBrowser) Stop() {
	r.Browser.Close()
}

func IsRuning(id int64) bool {
	return bs.BsManager.IsArride(id)
}

func (r *RunBrowser) Msg(msg string) {

}
