package runphone

import (
	"context"
	"tools/runtimes/scheduler"
)

type runphon struct {
	release   func()
	scheduler *scheduler.Runner
}

func New(fun func()) *runphon {
	return &runphon{
		release: fun,
	}
}
func (t *runphon) SetRunner(s *scheduler.Runner) {
	t.scheduler = s
}

func (t *runphon) Start(ctx context.Context) error {
	return nil
}

func (t *runphon) OnError(err error) {

}
func (t *runphon) OnClose() {
	if t.release != nil {
		t.release()
	}
}
func (t *runphon) OnChange(str string) {

}

func (t *runphon) Close() {

}
