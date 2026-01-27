package runhttp

import "context"

type runhttp struct {
}

func New() *runhttp {
	return &runhttp{}
}

func (t *runhttp) Start(ctx context.Context) error {
	return nil
}

func (t *runhttp) OnError(err error) {

}
func (t *runhttp) OnClose() {

}
func (t *runhttp) OnChange(str string) error {
	return nil
}
