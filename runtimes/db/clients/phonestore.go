package clients

import (
	"time"
	"tools/runtimes/apptask"

	"gorm.io/gorm"
)

type GormStore struct {
	db *gorm.DB
}

func NewGormTaskStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

func (s *GormStore) LoadTasks() ([]*apptask.AppTask, error) {
	// 1️⃣ 查数据库（只用 DB Model）
	var models []AppTask
	if err := s.db.Find(&models).Error; err != nil {
		return nil, err
	}

	// 2️⃣ 转换为 Domain Task
	tasks := make([]*apptask.AppTask, 0, len(models))
	for _, m := range models {
		task := &apptask.AppTask{
			Id:        m.Id,
			DeviceId:  m.DeviceId,
			Type:      m.Type,
			Data:      m.Data,
			Starttime: m.Starttime,
			Endtime:   m.Endtime,
			Cycle:     m.Cycle,
			Enabled:   m.Status == 1,
		}
		tasks = append(tasks, task)
	}

	// 3️⃣ 返回给 runtime（纯内存模型）
	return tasks, nil
}

func (s *GormStore) CreateRun(task *apptask.AppTask) (*apptask.AppTaskRun, error) {
	// 1️⃣ 先创建 DB Model
	model := &AppTaskRun{
		TaskId:    task.Id,
		RunStatus: 0,
		Exectime:  time.Now().Unix(),
	}

	// 2️⃣ 写数据库
	if err := s.db.Create(model).Error; err != nil {
		return nil, err
	}

	// 3️⃣ 转回 Domain Run
	return &apptask.AppTaskRun{
		RunId:     model.RunId,
		TaskId:    model.TaskId,
		Status:    model.RunStatus,
		StartTime: model.Exectime,
	}, nil
}

func (s *GormStore) FinishRun(runId int64, status int, msg string) error {
	return s.db.Model(&AppTaskRun{}).
		Where("run_id = ?", runId).
		Updates(map[string]any{
			"run_status": status,
			"msg":        msg,
			"donetime":   time.Now().Unix(),
		}).Error
}

func (s *GormStore) SaveTask(task *AppTask) error {
	if task.Id > 0 {
		return s.db.Model(&AppTask{}).
			Where("id = ?", task.Id).
			Updates(map[string]any{
				"type":      task.Type,
				"data":      task.Data,
				"device_id": task.DeviceId,
				"status":    task.Status,
				"starttime": task.Starttime,
				"endtime":   task.Endtime,
				"cycle":     task.Cycle,
			}).Error
	}

	if task.Addtime <= 0 {
		task.Addtime = time.Now().Unix()
	}

	return s.db.Create(task).Error
}

func (s *GormStore) DeleteTask(taskId int64) error {
	return s.db.Where("id = ?", taskId).Delete(&AppTask{}).Error
}

func (s *GormStore) SetTaskEnabled(taskId int64, enabled bool) error {
	status := 0
	if enabled {
		status = 1
	}

	return s.db.Model(&AppTask{}).
		Where("id = ?", taskId).
		Update("status", status).Error
}

func (s *GormStore) ForceStopRun(runId int64, operator string) error {
	return s.db.Model(&AppTaskRun{}).
		Where("run_id = ?", runId).
		Updates(map[string]any{
			"run_status": 2,
			"msg":        "force stop by " + operator,
			"donetime":   time.Now().Unix(),
		}).Error
}
