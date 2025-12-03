package suggestions

import (
	"tools/runtimes/db"
)

type SuggCate struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"index;not null"`
}
type Suggestion struct {
	Id           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Addtime      int64  `json:"addtime" gorm:"index;default:0"`               // 反馈时间
	Title        string `json:"title" gorm:"index;not null;type:varchar(40)"` // 标题
	Content      string `json:"content" gorm:"default:null"`                  // 反馈的详细内容
	CateId       int64  `json:"cate_id" gorm:"index;default:0"`               // 反馈的类别,0为未选择
	ReadTime     int64  `json:"read_time" gorm:"index;default:0"`             // 服务端阅读反馈的时间
	LastBackTime int64  `json:"last_back_time" gorm:"index;default:0"`        // 最后一次服务端反馈的时间
}
type SuggMessage struct { //对意见和建议内容进行再度讨论的内容
	Id      int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	SuggId  int64  `json:"sugg_id" gorm:"index;not null"`
	Addtime int64  `json:"addtime" gorm:"index;default:0"`
	Content string `json:"content" gorm:"not null"`                    // 内容
	Rule    int    `json:"rule" gorm:"type:tinyint(1);index;not null"` // 内容角色,0为客户端,1为服务端回答
}

func init() {
	db.DB.AutoMigrate(&SuggCate{})
	db.DB.AutoMigrate(&Suggestion{})
	db.DB.AutoMigrate(&SuggMessage{})
}
