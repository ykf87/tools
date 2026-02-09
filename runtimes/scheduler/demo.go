package scheduler

import (
	"context"
	"fmt"
	"time"
)

func Test() {
	s := New(context.Background())
	rr := s.NewRunner(func(ctx context.Context) error {
		fmt.Println("rrrr---")
		time.Sleep(time.Second * 3)
		return nil
	}, 0, nil).SetMaxTry(3)
	nn := s.NewRunner(func(ctx context.Context) error {
		fmt.Println("nnnn--")
		return fmt.Errorf("nnn error")
	}, 0, nil)
	fmt.Println(rr.GetID(), nn.GetID())
	rr.SetError(func(err error, tried int32) {
		fmt.Println("执行失败回调:", err, ",重试次数:", tried)
	})
	rr.SetCloser(func() {
		fmt.Println("rr任务被关闭了")
	})
	nn.SetCloser(func() {
		fmt.Println("nn 任务被关闭")
	})
	nn.SetError(func(err error, tried int32) {
		fmt.Println("---nnn 错误信息:", err, ",重试次数:", tried)
	})
	rr.Every(time.Second * 2).Run()
	nn.SetMaxTry(2).RunNow()
	go func() {
		time.Sleep(time.Second * 11)
		rr.Stop()
	}()

}
