package tasklog

// Task 并不是写入数据库的,而是作为任务载体发送到websocket中
// 用于
type Task struct {
	TaskID  string `json:"task_id"`  // 前端使用此字段作为判断,TaskID在每次发起任务时通过调用方设定
	Title   string `json:"title"`    // 任务名称
	BeginAt int64  `json:"begin_at"` // 开始于什么时间
}

type TaskLog struct {
	ID     int64  `json:"id" gorm:"primaryKey;autoIncrement"`  // 表自增id
	TaskID int64  `json:"task_id" gorm:"index;default:0"`      // 任务对于的表的id号,0代表任务并不是从数据库发起的
	Type   string `json:"type" gorm:"index;not null"`          // 任务类型,比如来自任务还是来自用户信息自动同步
	Title  string `json:"title" gorm:"type:varchar(32);index"` // 任务名称
	UUID   string `json:"uuid" gorm:"index;not null"`          //
}
