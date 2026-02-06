package runner

import (
	"errors"
	"tools/runtimes/runner/runbrowser"
)

type Runner interface {
	Start() error
	Stop()
}

func GetRunner(tp int, opt any) (Runner, error) {
	switch tp {
	case 0: // 浏览器
		return runbrowser.New(opt, true)
	case 1: // 手机
	case 2: // http
	}
	return nil, errors.New("找不到该类型的执行器")
}

func IsRuning(tp int, id int64) bool {
	switch tp {
	case 0: // 浏览器
		return runbrowser.IsRuning(id)
	case 1: // 手机
	case 2: // http
	}
	return false
}
