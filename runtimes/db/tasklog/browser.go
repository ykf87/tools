package tasklog

import (
	"context"
	"errors"
	"fmt"
	"time"
	"tools/runtimes/bs"
)

func (tr *TaskRunner) callweb(ctx context.Context) error {
	opt, ok := tr.opt.(*bs.Options)
	if !ok {
		return fmt.Errorf("option error")
	}
	opt.Ctx = ctx
	bss, err := bs.BsManager.New(opt.ID, opt, true)
	if err != nil {
		return err
	}

	bss.Opts.Msg = make(chan string)
	tr.bss = bss

	if err := tr.bss.OpenBrowser(); err != nil {
		return err
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
