package task

import (
	"errors"
	"time"
)

type TaskExecutor struct {
	Body      *TaskBody
	Instance  *TaskInstance
	Reporter  Reporter
	CreateSub func(taskID int64, name string) *SubTaskInstance
}

func (e *TaskExecutor) Execute() error {
	if e.Body == nil || e.Instance == nil {
		return errors.New("task body or instance is nil")
	}

	r := safeReporter(e.Reporter)

	e.startTask(r)

	total := e.Body.TotalWeight()
	var done float64

	for _, sub := range e.Body.SubTasks {
		subIns := e.CreateSub(e.Instance.ID, sub.Name())
		e.startSub(subIns, r)

		ctx := &SubTaskContext{
			TaskID:    e.Instance.ID,
			SubTaskID: subIns.ID,
			Reporter:  r,
		}

		if err := sub.Execute(ctx); err != nil {
			e.failSub(subIns, err, r)
			e.failTask(err, r)
			return err // 交给 Runner
		}

		e.successSub(subIns, r)

		done += sub.Weight()
		e.Instance.Progress = done / total * 100
		r.Progress(e.Instance.ID, e.Instance.Progress)
	}

	e.successTask(r)
	return nil
}

func (e *TaskExecutor) startTask(r Reporter) {
	e.Instance.Status = StatusRunning
	e.Instance.StartAt = time.Now()
	r.Log(e.Instance.ID, "任务开始")
}

func (e *TaskExecutor) successTask(r Reporter) {
	now := time.Now()
	e.Instance.Status = StatusSuccess
	e.Instance.EndAt = &now
	e.Instance.Progress = 100
	r.Log(e.Instance.ID, "任务完成")
}

func (e *TaskExecutor) failTask(err error, r Reporter) {
	now := time.Now()
	e.Instance.Status = StatusFailed
	e.Instance.EndAt = &now
	e.Instance.Error = err.Error()
	r.Log(e.Instance.ID, "任务失败："+err.Error())
}

func (e *TaskExecutor) startSub(s *SubTaskInstance, r Reporter) {
	s.Status = StatusRunning
	s.StartAt = time.Now()
	r.Log(e.Instance.ID, "子任务开始："+s.Name)
}

func (e *TaskExecutor) successSub(s *SubTaskInstance, r Reporter) {
	now := time.Now()
	s.Status = StatusSuccess
	s.EndAt = &now
}

func (e *TaskExecutor) failSub(s *SubTaskInstance, err error, r Reporter) {
	now := time.Now()
	s.Status = StatusFailed
	s.EndAt = &now
	s.Error = err.Error()
}
