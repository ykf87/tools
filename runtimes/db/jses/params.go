package jses

import (
	"fmt"
	"tools/runtimes/db"
)

// js内容和变量对应表
// 此表仅是定规则,并不存储真实数据
// 仅用作在编辑任务时的一些限制和数据预填
type JsParam struct {
	JsID         int64   `json:"js_id" gorm:"not null;primaryKey"`             // Js 表ID
	CodeName     string  `json:"code_name" gorm:"not null;primaryKey"`         // 用于替换Js表的Content中的变量名称
	Type         string  `json:"type" gorm:"index;type:varchar(30); not null"` // 类型
	Label        string  `json:"label" gorm:"type:varchar(32);not null"`       // 展示的名称
	Required     int     `json:"required" gorm:"type:tinyint(1);default:0"`    // 是否必须
	Placeholder  string  `json:"placeholder" gorm:"type:varchar(150);"`        // 提示
	Tips         string  `json:"tips"`                                         // 长提示,类似说明
	Options      string  `json:"options" parse:"json"`                         // 默认的选项,json格式
	DownDataType *int    `json:"down_data_type"`                               // 下拉菜单的数据类型,0为预设值,1为api接口
	Mulit        int     `json:"mulit" gorm:"default:0"`                       // 是否可多选,针对下拉菜单的
	DefaultValue string  `json:"default"`                                      // 默认值
	Rules        string  `json:"rules"`                                        // 验证规则,js的函数,并且传入输入数据,返回true为验证通过,其他返回值均视为失败.返回string为错误信息提示
	Api          string  `json:"api"`                                          // 数据接口,需要是此后台支持的.此接口仅用于生成一些预设值,也就是选项的值.和task_params表的接口作用不一样
	Method       string  `json:"method"`                                       // 接口请求的方式
	ApiParams    *string `json:"api_params"`                                   // 数据接口调用时的参数
	ApiRespFun   *string `json:"api_resp_fun" gorm:"type:longtext"`            // api接口返回的结果使用这个js函数重构
}

func (this *Js) GenParams() error {
	db.DB.Where("js_id = ?", this.ID).Delete(&JsParam{})

	if len(this.Params) > 0 {
		for _, v := range this.Params {
			if v.Label == "" {
				return fmt.Errorf("%s 的 参数名称 未设置", v.CodeName)
			}
			if v.Type == "" {
				v.Type = "input"
			}
			v.JsID = this.ID
		}
		return db.DB.Create(this.Params).Error
	}
	return nil
}

func GetParamsByJsID(id any) []*JsParam {
	var lst []*JsParam
	db.DB.Model(&JsParam{}).Where("js_id = ?", id).Order("required DESC").Find(&lst)
	return lst
}
