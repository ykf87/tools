package runweb

import (
	"context"
	"time"
)

type runweb struct {
}

func New() *runweb {
	return &runweb{}
}

func (t *runweb) Start(ctx context.Context) error {
	time.Sleep(time.Second * 3)
	return nil
}

func (t *runweb) OnError(err error) {

}
func (t *runweb) OnClose() {

}
func (t *runweb) OnChange(str string) error {
	return nil
}
