package jses

// js内容和变量对应表
// 此表仅是定规则,并不存储真实数据
type JsParam struct {
	JsID         int64  `json:"js_id" gorm:"not null;primaryKey"`             // Js 表ID
	CodeName     string `json:"code_name" gorm:"not null;primaryKey"`         // 用于替换Js表的Content中的变量名称
	Type         string `json:"type" gorm:"index;type:varchar(30); not null"` // 类型
	Label        string `json:"label" gorm:"type:varchar(32);not null"`       // 展示的名称
	Required     int    `json:"required" gorm:"type:tinyint(1);default:0"`    // 是否必须
	Placeholder  string `json:"placeholder" gorm:"type:varchar(150);"`        // 提示
	Tips         string `json:"tips"`                                         // 长提示,类似说明
	Options      string `json:"options" parse:"json"`                         // 默认的选项,json格式
	DownDataType int    `json:"down_data_type" gorm:"default:null"`           // 下拉菜单的数据类型,0为预设值,1为api接口
	Mulit        int    `json:"mulit" gorm:"default:0"`                       // 是否可多选,针对下拉菜单的
	Default      string `json:"default"`                                      // 默认值
	Rules        string `json:"rules"`                                        // 验证规则
	Api          string `json:"api"`                                          // 数据接口,需要是此后台支持的.此接口仅用于生成一些预设值,也就是选项的值.和task_params表的接口作用不一样
	Method       string `json:"method"`                                       // 接口请求的方式
	ApiParams    string `json:"api_params"`                                   // 数据接口调用时的参数
}

type TypesStruct struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

var Types = []TypesStruct{
	TypesStruct{Key: "input", Name: "文本"},
	TypesStruct{Key: "input-number", Name: "数字"},
	TypesStruct{Key: "textarea", Name: "长内容"},
	TypesStruct{Key: "select", Name: "下拉选择"},
	TypesStruct{Key: "radio", Name: "单选"},
	TypesStruct{Key: "checkbox", Name: "多选"},
}

func (this *Js) GenParams() {

}
