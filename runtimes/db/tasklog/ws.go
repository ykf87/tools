package tasklog

import (
	"fmt"
	"math"
	"tools/runtimes/config"
	"tools/runtimes/listens/ws"
)

type TaskWs struct {
	TaskID  string          `json:"task_id"`
	Title   string          `json:"title"`
	BeginAt int64           `json:"begin"`
	MaxCh   int             `json:"max_ch"`
	Runner  []*TaskRunnerWs `json:"runner"`
	Runing  bool            `json:"runing"`
	IsTemp  bool            `json:"is_temp"` // 是否是临时任务
}

type TaskRunnerWs struct {
	ID       string  `json:"id"`
	TaskID   string  `json:"task_id"`
	Title    string  `json:"title"`
	Msg      string  `json:"msg"`
	ErrMsg   string  `json:"err_msg"`
	StartAt  int64   `json:"start_at"`
	EndAt    int64   `json:"end_at"`
	RunTimes int     `json:"run_times"` // 已执行次数, 周期任务
	Total    float64 `json:"total"`     // 执行总量,比如下载文件总大小,或者执行养号总数
	Doned    float64 `json:"doned"`     // 已执行次数或已完成数量
	Percent  float64 `json:"percent"`   // 执行百分比
}

// 发送任务
func (t *Task) Sent() *TaskWs {
	tws := &TaskWs{
		TaskID:  t.TaskID,
		Title:   t.Title,
		BeginAt: t.BeginAt,
		MaxCh:   t.MaxCh,
		Runner:  []*TaskRunnerWs{},
		IsTemp:  t.IsTemp,
		Runing:  true,
	}

	if bt, err := config.Json.Marshal(map[string]any{
		"type": "task",
		"data": tws,
	}); err == nil {
		ws.Broadcost(bt)
		return tws
	}
	return nil
}

// ws连上后主动发送一次在执行的task任务，此处仅需发送任务下的执行内容,无需发送整个task
func (v *TaskRunner) newWs(msg, errmsg string) ([]byte, error) {
	if v.Runner == nil {
		return nil, fmt.Errorf("执行器未创建")
	}
	var per float64
	if v.Total > 0 {
		per = math.Round(v.Doned/v.Total*100*100) / 100
	}
	trws := &TaskRunnerWs{
		ID:       v.ID,
		TaskID:   v.taskID,
		Title:    v.Title,
		Msg:      msg,
		ErrMsg:   errmsg,
		StartAt:  v.Runner.GetStartAt().Unix(),
		EndAt:    v.endAt,
		RunTimes: v.Runner.GetRunTimes(),
		Total:    v.Total,
		Doned:    v.Doned,
		Percent:  per,
	}
	return config.Json.Marshal(map[string]any{
		"type": "taskrunner",
		"data": trws,
	})
}

// 发送执行任务
func (tr *TaskRunner) Sent(msg string) {
	if bt, err := tr.newWs(msg, ""); err == nil {
		ws.Broadcost(bt)
	}
}

func (tr *TaskRunner) SentErr(msg string) {
	if bt, err := tr.newWs("", msg); err == nil {
		ws.Broadcost(bt)
	}
}
