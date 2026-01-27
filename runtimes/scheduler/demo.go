package scheduler

import (
	"context"
	"fmt"
	"time"
)

func Test() {
	s := New()
	rr := s.NewRunner(func(ctx context.Context) error {
		fmt.Println("rrrr---")
		time.Sleep(time.Second * 3)
		return nil
	}, 0).SetMaxTry(3)
	nn := s.NewRunner(func(ctx context.Context) error {
		fmt.Println("nnnn--")
		return fmt.Errorf("nnn error")
	}, 0)
	fmt.Println(rr.GetID(), nn.GetID())
	rr.SetError(func(err error) {
		fmt.Println("执行失败回调:", err)
	})
	rr.SetCloser(func() {
		fmt.Println("rr任务被关闭了")
	})
	nn.SetCloser(func() {
		fmt.Println("nn 任务被关闭")
	})
	nn.SetError(func(err error) {
		fmt.Println("---nnn 错误信息:", err)
	})
	rr.Every(time.Second * 2).Run()
	nn.SetMaxTry(2).RunNow()
	go func() {
		time.Sleep(time.Second * 11)
		rr.Stop()
	}()

}
