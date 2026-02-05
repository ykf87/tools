package tasklog

import (
	"context"
	"errors"
	"fmt"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/logs"
)

func (tr *TaskRunner) SetMsg(msg string) {
	if tr.Msg != nil {
		tr.Msg <- msg
	}
}
func (tr *TaskRunner) SetErrMsg(msg string) {
	if tr.ErrMsg != nil {
		tr.ErrMsg <- msg
	}
}

func (tr *TaskRunner) callweb(ctx context.Context) error {
	opt, ok := tr.opt.(*bs.Options)
	if !ok {
		// tr.SetErrMsg("配置错误,任务未开始执行")
		return fmt.Errorf("配置错误,任务未开始执行")
	}

	opt.Ctx = ctx
	bss, err := bs.BsManager.New(opt.ID, opt, true)
	if err != nil {
		// tr.SetErrMsg("浏览器创建失败:" + err.Error())
		return fmt.Errorf("浏览器创建失败: %s", err.Error())
	}

	if bss.Opts.Proxy == "" && bss.Opts.Pc != nil {
		if _, err := bss.Opts.Pc.Run(false); err == nil {
			bss.Opts.Proxy = bss.Opts.Pc.Listened()
		}
	}
	if bss.Opts.Proxy != "" {
		tr.ProxyUrl = bss.Opts.Proxy
	}
	tr.SetMsg("任务开始,获取信息中...")

	bss.Opts.Msg = make(chan string)
	tr.Bss = bss

	if err := tr.Bss.OpenBrowser(); err != nil {
		// tr.SetErrMsg("浏览器打开失败:" + err.Error())
		return fmt.Errorf("浏览器打开失败: %s", err.Error())
	}

	ctxx := tr.Runner.GetCtx()
	for {
		select {
		case msg := <-bss.Opts.Msg:
			if err := tr.callBack(msg, tr); err != nil {
				return err
			}
		case <-ctxx.Done():
			if errors.Is(ctxx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("超时退出")
			}
			// if errors.Is(ctxx.Err(), context.Canceled) {
			// 	return fmt.Errorf("被主动关闭")
			// }
			goto CLOSER
		case <-bss.GetCtx().Done():
			goto CLOSER
		}
	}
CLOSER:
	bs.BsManager.Remove(opt.ID)
	return nil
}

// 启动执行任务
func (tr *TaskRunner) StartWeb(opt *bs.Options, timeout time.Duration) error {
	tr.opt = opt

	tr.Runner = tr.sch.NewRunner(tr.callweb, timeout, tr.ctx)
	tr.Runner.SetError(func(err error) {
		select {
		case tr.ErrMsg <- err.Error():
			logs.Error("taskRunner StartWeb error:" + err.Error())
		case <-tr.ctx.Done():
			return
		}
	})
	tr.Runner.SetCloser(func() {
		tr.endAt = time.Now().Unix()
		go tr.Sent("本次任务结束")
		tr.cancle()
	})

	return nil
}
