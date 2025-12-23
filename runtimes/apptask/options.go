package apptask

import "time"

type Options struct {
	TickInterval time.Duration // 内存扫描周期
	Executor     Executor
	Store        Store
}
