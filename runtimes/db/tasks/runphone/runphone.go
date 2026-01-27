package runphone

import "context"

type runphon struct {
}

func New() *runphon {
	return &runphon{}
}

func (t *runphon) Start(ctx context.Context) error {
	return nil
}

func (t *runphon) OnError(err error) {

}
func (t *runphon) OnClose() {

}
func (t *runphon) OnChange(str string) error {
	return nil
}
