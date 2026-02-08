package task

type SubTask interface {
	Name() string
	Weight() float64
	Execute(ctx *SubTaskContext) error
}

type SubTaskContext struct {
	TaskID    int64
	SubTaskID int64
	Reporter  Reporter
}
