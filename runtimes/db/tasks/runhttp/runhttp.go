package runhttp

import (
	"context"
	"tools/runtimes/scheduler"
)

type runhttp struct {
	release   func()
	scheduler *scheduler.Runner
}

func New(fun func()) *runhttp {
	return &runhttp{
		release: fun,
	}
}
func (t *runhttp) SetRunner(s *scheduler.Runner) {
	t.scheduler = s
}

func (t *runhttp) Start(ctx context.Context) error {
	return nil
}

func (t *runhttp) OnError(err error) {

}
func (t *runhttp) OnClose() {
	if t.release != nil {
		t.release()
	}
}
func (t *runhttp) OnChange(str string) error {
	return nil
}
