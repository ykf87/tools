package browser

import "time"

type Options struct {
	ExecPath  string
	UserDir   string
	Proxy     string
	UserAgent string
	Headless  bool
	Width     int
	Height    int
	Timeout   time.Duration
	Temp      bool // 是否临时浏览器
}
