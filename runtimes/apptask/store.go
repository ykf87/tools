package apptask

type Store interface {
	LoadPendingTasks() ([]*AppTask, error)
	CreateRun(task *AppTask) (*AppTaskMsg, error)
	FinishRun(runId int64, status int, msg string) error
}
