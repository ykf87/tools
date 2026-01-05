// 执行的js脚本,浏览器和手机端都是使用js来执行
package jses

import "tools/runtimes/db"

// js脚本表,content是脚本的内容
// replace_prev 默认是 <<
// replace_end 默认是 >>
type Js struct {
	ID          int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Code        string `json:"code" gorm:"uniqueIndex;not null"`              // 唯一标识符
	Name        string `json:"name" gorm:"index"`                             // 名称
	IsSys       int    `json:"is_sys" gorm:"type:tinyint(1);index;default:1"` // 是否是从服务端获取的脚本,如果从服务器获取的脚本,将使用aes加密
	AdminID     int64  `json:"admin_id" gorm:"index;default:0"`               // 管理员id, 如果是系统的则为0,如果是用户自己写的,则对应用户的id
	Content     string `json:"content" gorm:"not null;type:longtext"`         // 执行的脚本
	ReplacePrev string `json:"replace_prev"`                                  // 变量替换前缀
	ReplaceEnd  string `json:"replace_end"`                                   // 变量替换后缀
	Icon        string `json:"icon"`                                          // 此js的图标
}

// js内容和变量对应表
// 此表仅是定规则,并不存储真实数据
type JsParam struct {
	JsID        int64  `json:"js_id" gorm:"not null;primaryKey"`             // Js 表ID
	CodeName    string `json:"code_name" gorm:"not null;primaryKey"`         // 用于替换Js表的Content中的变量名称
	Type        string `json:"type" gorm:"index;type:varchar(30); not null"` // 类型
	Label       string `json:"label" gorm:"type:varchar(32);not null"`       // 展示的名称
	Required    int    `json:"required" gorm:"type:tinyint(1);default:0"`    // 是否必须
	Placeholder string `json:"placeholder" gorm:"type:varchar(150);"`        // 提示
	Tips        string `json:"tips"`                                         // 长提示,类似说明
	Options     string `json:"options" parse:"json"`                         // 默认的选项,json格式
	Default     string `json:"default"`                                      // 默认值
	Rules       string `json:"rules"`                                        // 验证规则
	Api         string `json:"api"`                                          // 数据接口,需要是此后台支持的.此接口仅用于生成一些预设值,也就是选项的值.和task_params表的接口作用不一样
	Method      string `json:"method"`                                       // 接口请求的方式
	ApiParams   string `json:"api_params"`                                   // 数据接口调用时的参数
}

func init() {
	db.DB.AutoMigrate(&Js{})
	db.DB.AutoMigrate(&JsParam{})
}

// 根据脚本ID获取脚本
func GetJsById(id int64) *Js {
	if id < 1 {
		return nil
	}
	jsobj := new(Js)
	db.DB.Model(&Js{}).Where("id = ?", id).First(jsobj)
	return jsobj
}
