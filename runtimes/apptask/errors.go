package apptask

import "errors"

var (
	// 任务不存在
	ErrTaskNotFound = errors.New("task not found")

	// 任务存在但未在运行
	ErrTaskNotRunning = errors.New("task is not running")

	// 任务正在运行
	ErrTaskRunning = errors.New("task is running")

	// 调度器未启动
	ErrManagerNotStarted = errors.New("manager not started")

	// 任务已存在（重复注册）
	ErrTaskExists = errors.New("task already exists")

	// 投递失败
	ErrDeliveryFailed = errors.New("task delivery failed")

	// 运行记录不存在
	ErrRunNotFound = errors.New("task run not found")
)
