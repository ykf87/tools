package proxys

type Subscribe struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"not null;uniqueIndex;"` // 自己备注的名称,必填且不重复
	Url  string `json:"url" gorm:"not null;uniqueIndex"`   // 订阅地址,不能重复订阅
}
