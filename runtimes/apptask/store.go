package apptask

type Store interface {
	// 系统启动恢复
	LoadTasks() ([]*AppTask, error)

	// 任务执行记录
	CreateRun(task *AppTask) (*AppTaskRun, error)
	FinishRun(runId int64, status int, msg string) error

	// // 后台管理
	// SaveTask(task *AppTask) error
	// DeleteTask(taskId int64) error
	// SetTaskEnabled(taskId int64, enabled bool) error

	// // 可选（推荐）
	// ForceStopRun(runId int64, operator string) error
}
