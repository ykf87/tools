package apptask

type Delivery interface {
	Mode() string // "ws" | "api"

	// 调度器调用
	Deliver(task *AppTask) error

	// API 模式下由 HTTP 调用
	Pick(deviceId string) *AppTask
}
