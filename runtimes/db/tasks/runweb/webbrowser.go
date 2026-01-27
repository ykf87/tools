package runweb

import (
	"context"
	"fmt"
	"time"
	"tools/runtimes/scheduler"
)

type runweb struct {
	release   func()
	scheduler *scheduler.Runner
}

func New(donfun func()) *runweb {
	return &runweb{
		release: donfun,
	}
}

func (t *runweb) SetRunner(s *scheduler.Runner) {
	t.scheduler = s
}

func (t *runweb) Start(ctx context.Context) error {
	time.Sleep(time.Second * 3)
	fmt.Println("执行完成!", t.scheduler.GetRunTimes(), t.scheduler.GetSigleRunTime(), t.scheduler.GetTotalTime())
	return fmt.Errorf("---")
}

func (t *runweb) OnError(err error) {

}
func (t *runweb) OnClose() {
	if t.release != nil {
		t.release()
	}
	fmt.Println("任务结束,总执行时间:", t.scheduler.GetTotalTime(), ".重试:", t.scheduler.GetTryTimers())
}
func (t *runweb) OnChange(str string) error {
	return nil
}
