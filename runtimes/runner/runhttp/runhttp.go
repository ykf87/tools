package runhttp

import (
	"errors"
	"fmt"
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
func (p *Httpp) Start(
	timeout time.Duration,
	callback func(msg, data string) error,
	errcallback func(msg string),
	msgcallback func(msg string),
) error {
	return nil
}

func (p *Httpp) Msg(msg string) {
	fmt.Println(msg)
}

// 停止
func (p *Httpp) Stop() {

}

// 是否启动
func IsRuning(id int64) bool {
	return true
}
