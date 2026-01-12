package db

type Sys struct {
	Id int64 `json:"id" gorm:"primaryKey;autoIncrement"`
}

type TypesStruct struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

var Types = []TypesStruct{
	TypesStruct{Key: "input", Name: "文本"},
	TypesStruct{Key: "input-number", Name: "数字"},
	TypesStruct{Key: "textarea", Name: "长内容"},
	TypesStruct{Key: "select", Name: "选择"},
	TypesStruct{Key: "medias", Name: "媒体文件"}, // 媒体文件的选择应该在任务编辑中添加
	// TypesStruct{Key: "api", Name: "API获取"},
}
