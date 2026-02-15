package runhttp

import (
	"errors"
	"time"
)

type Option struct {
}

type Httpp struct {
	opt *Option
}

func New(opt any) (*Httpp, error) {
	if o, ok := opt.(*Option); ok {
		return &Httpp{
			opt: o,
		}, nil
	}
	return nil, errors.New("option is error")
}

// 启动
func (p *Httpp) Start(timeout time.Duration, callback func(str string) error) error {
	return nil
}

// 停止
func (p *Httpp) Stop() {

}

// 是否启动
func IsRuning(id int64) bool {
	return true
}
