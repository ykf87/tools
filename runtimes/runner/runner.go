// 封装3中获取方式,0使用指纹浏览器打开, 1使用手机端打开, 2使用golang的http执行
package runner

import (
	"context"
	"errors"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/config"
	"tools/runtimes/proxy"
	"tools/runtimes/runner/runbrowser"
	"tools/runtimes/runner/runhttp"
	"tools/runtimes/runner/runphone"
)

type Runner interface {
	Start(time.Duration, func(string) error) error
	Stop()
}

// 生成浏览器的配置,目的是为了配置的一致性
func GenWebOpt(
	ctx context.Context, // 上下文
	id int64, // 浏览器id
	headless bool, // 是否隐藏执行,true 不打开窗口, false 显示窗口
	url string, // 默认打开的地址
	js string, // 执行的js脚本
	proxyConfig *proxy.ProxyConfig, // 必须使用本地启动的, http://127.0.0.1:3521 这种地址也可以通过 proxy.Client() 生成
	timeout time.Duration, // 超时
	width int, // 浏览器宽度
	height int, // 浏览器高度
	language string, // 浏览器使用的语言
	timezone string, // 浏览器使用的时区
) *bs.Options {
	var whmap map[string]int
	config.AdminWidthAndHeight.Range(func(k, v any) bool {
		if whmap == nil {
			if vv, ok := v.(map[string]int); ok {
				whmap = vv
			}
		}
		return true
	})

	if width < 1 && whmap != nil {
		if w, ok := whmap["widht"]; ok {
			width = w
		}
	}
	if height < 1 && whmap != nil {
		if h, ok := whmap["height"]; ok {
			height = h
		}
	}
	opt := &bs.Options{
		Width:    width,
		Height:   height,
		Headless: headless,
		ID:       id,
		Url:      url,
		JsStr:    js,
		Timezone: timezone,
		Language: language,
		Ctx:      ctx,
		Timeout:  timeout,
		Pc:       proxyConfig,
	}
	return opt
}

// 生成手机端的配置
func GenPhoneOpt() *runphone.Option {
	return &runphone.Option{}
}

// 生成http端的配置
func GenHttpOpt() *runhttp.Option {
	return &runhttp.Option{}
}

func GetRunner(tp int, opt any) (Runner, error) {
	switch tp {
	case 0: // 浏览器
		return runbrowser.New(opt, true)
	case 1: // 手机
		return runphone.New(opt)
	case 2: // http
		return runhttp.New(opt)
	}
	return nil, errors.New("找不到该类型的执行器")
}

func IsRuning(tp int, id int64) bool {
	switch tp {
	case 0: // 浏览器
		return runbrowser.IsRuning(id)
	case 1: // 手机
		return runphone.IsRuning(id)
	case 2: // http
		return runhttp.IsRuning(id)
	}
	return false
}
