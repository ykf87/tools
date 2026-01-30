package tasks

import (
	"fmt"
	"tools/runtimes/scheduler"
)

// 任务设备表
type TaskClients struct {
	TaskID     int64             `json:"task_id" gorm:"primaryKey"`     // 任务ID
	DeviceType int               `json:"device_type" gorm:"primaryKey"` // 设备类型, 0-web端 1-手机端 2-发起http请求
	DeviceID   int64             `json:"device_id" gorm:"primaryKey"`   // 设备id,在clients下的 browsers 或者 phones 表
	tsk        *Task             `json:"-" gorm:"-"`                    // 任务
	runner     *scheduler.Runner `json:"-" gorm:"-"`                    // 调度器中的任务
}

func (t *Task) GetClients() []*TaskClients {
	var tcs []*TaskClients
	if t.ID > 0 {
		dbs.Model(&TaskClients{}).Where("task_id = ?", t.ID).Find(&tcs)
	}
	return tcs
}

func (t *TaskClients) GetName() string {
	return fmt.Sprintf("%d%d%d", t.TaskID, t.DeviceType, t.DeviceID)
}
