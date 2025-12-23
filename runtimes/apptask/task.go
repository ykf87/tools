package apptask

type AppTask struct {
	Id        int64
	Type      string
	Data      string
	DeviceId  string
	Addtime   int64
	Starttime int64
	Endtime   int64
	Cycle     int64
	Enabled   bool // 后台启停
	Timeout   int64
}

type AppTaskRun struct {
	RunId     int64
	TaskId    int64
	Status    int // 0=running 1=success 2=fail
	Msg       string
	StartTime int64
	EndTime   int64
}
