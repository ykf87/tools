package runtask

import "sync"

// TaskLog 任务执行详细日志表
type TaskLog struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID    int64  `json:"task_id" gorm:"index;not null"`
	TaskRunID int64  `json:"task_run_id" gorm:"index;not null"`
	Addtime   int64  `json:"addtime" gorm:"index;not null"`
	LogStatus int    `json:"log_status" gorm:"index;default:0"`
	Msg       string `json:"msg" gorm:"default:null"`
}

type RunTask struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID string `json:"uuid" gorm:"not null;"`
	mu   sync.RWMutex
	wake chan struct{}
}
