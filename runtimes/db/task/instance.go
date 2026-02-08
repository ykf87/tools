package task

import "time"

type TaskInstance struct {
	ID       int64
	TaskCode string
	Status   Status
	Progress float64
	StartAt  time.Time
	EndAt    *time.Time
	Error    string
}

type SubTaskInstance struct {
	ID      int64
	TaskID  int64
	Name    string
	Status  Status
	StartAt time.Time
	EndAt   *time.Time
	Error   string
}
