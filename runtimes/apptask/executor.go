package apptask

// Executor 只负责：把任务“发给设备”
type Executor interface {
	Execute(task *AppTask) error
}
