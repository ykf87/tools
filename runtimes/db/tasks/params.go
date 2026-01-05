package tasks

// 任务对应的js使用的参数的值
// 或者他的数据调用方式
type TaskParam struct {
	TaskID   int64  `json:"task_id" gorm:"primaryKey;not null"` // 任务表的id
	JsID     int64  `json:"js_id" gorm:"primaryKey;not null"`   // js表的id
	CodeName string `json:"code_name" gorm:"not null;"`         // 替换js内容的键
	Value    string `json:"value" gorm:"default:null"`          // 数据
	Api      string `json:"api"`                                // 获取数据的接口,value和api两者需要只是存在一个,value优先级高于api
	Method   string `json:"method"`                             // 数据获取接口的调用方式
	Params   string `json:"params"`                             // 获取数据时的参数
}
