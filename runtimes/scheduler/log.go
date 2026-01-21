package scheduler

import "time"

type ExecLog struct {
	TaskID     string
	StartTime  time.Time
	EndTime    time.Time
	Duration   time.Duration
	Success    bool
	Error      string
	RetryCount int
	Timeout    bool
}
