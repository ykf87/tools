package runphone

import (
	"errors"
	"time"
)

type Option struct {
}

type Phone struct {
	opt *Option
}

func New(opt any) (*Phone, error) {
	if o, ok := opt.(*Option); ok {
		return &Phone{
			opt: o,
		}, nil
	}
	return nil, errors.New("option is error")
}

// 启动
func (p *Phone) Start(timeout time.Duration, callback func(str string) error) error {
	return nil
}

// 停止
func (p *Phone) Stop() {

}

// 是否启动
func IsRuning(id int64) bool {
	return true
}
