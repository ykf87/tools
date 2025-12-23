package apptask

// 任务定义（静态）
type AppTask struct {
	Id        int64
	Type      string
	Data      string
	DeviceId  string
	Addtime   int64
	Starttime int64
	Endtime   int64
	Cycle     int64 // 秒，0 = 不重复
}

// 任务执行记录（一次执行 = 一条）
type AppTaskMsg struct {
	RunId     int64
	TaskId    int64
	RunStatus int // 0=执行中 1=成功 2=失败
	Msg       string
	Exectime  int64
	Donetime  int64
}
