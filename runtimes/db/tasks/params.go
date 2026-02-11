package tasks

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 任务对应的js使用的参数的值
// 或者他的数据调用方式
type TaskParam struct {
	TaskID   int64  `json:"task_id" gorm:"primaryKey;not null"`      // 任务表的id
	JsID     int64  `json:"js_id" gorm:"primaryKey;not null"`        // js表的id
	CodeName string `json:"code_name" gorm:"primaryKey;not null;"`   // 替换js内容的键
	Value    any    `json:"value" gorm:"default:null;type:longtext"` // 数据
}

// 当前拥有的任务参数
func (this *Task) GetParams() []*TaskParam {
	Dbs.DB().Model(&TaskParam{}).Where("task_id = ?", this.ID).Find(&this.Params)
	return this.Params
}

// 设置任务参数
func (this *Task) GenParams(ps []*TaskParam) error {
	if this.ID > 0 {
		Dbs.Write(func(tx *gorm.DB) error {
			return tx.Where("task_id = ?", this.ID).Delete(&TaskParam{}).Error
		})
		// dbs.DB().Where("task_id = ?", this.ID).Delete(&TaskParam{})
	}

	var pss []*TaskParam
	for _, v := range ps {
		v.TaskID = this.ID
		v.JsID = this.Script
		pss = append(pss, v)
		// switch val := v.Value.(type) {
		// case string:
		// 	if val != "" {
		// 		pss = append(pss, v)
		// 	}
		// case int, int64, float64:
		// 	if val != 0 {
		// 		pss = append(pss, v)
		// 	}
		// default:
		// 	v.Value = ""
		// 	pss = append(pss, v)
		// }
	}

	if len(pss) > 0 {
		return Dbs.Write(func(tx *gorm.DB) error {
			return tx.Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&pss).Debug().Error
		})
		// return dbs.DB().
		// 	Clauses(clause.OnConflict{
		// 		DoNothing: true,
		// 	}).
		// 	Create(&pss).Debug().Error
	}

	return nil
}
