// 执行的js脚本,浏览器和手机端都是使用js来执行
package jses

type Js struct {
	ID      int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name    string `json:"name" gorm:"index"`
	IsSys   int    `json:"is_sys" gorm:"type:tinyint(1);index;default:1"` // 是否是从服务端获取的脚本
	Content string `json:"content" gorm:"not null;type:longtext"`         // 执行的脚本
}
