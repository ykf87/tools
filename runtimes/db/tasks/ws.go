package tasks

// import (
// 	"tools/runtimes/config"
// 	"tools/runtimes/listens/ws"
// )

// type WsTask struct {
// 	ID      int64           `json:"id"`
// 	UUID    string          `json:"uuid"`
// 	IsRun   bool            `json:"is_run"`
// 	Title   string          `json:"title"`
// 	Tags    []string        `json:"tags"`
// 	ErrMsg  string          `json:"err_msg"`
// 	Msg     string          `json:"msg"`
// 	Runners []*WsTaskRuning `json:"runners"`
// }
// type WsTaskRuning struct {
// 	UUID      string `json:"uuid"`
// 	StartAt   int64  `json:"start_at"`
// 	EndAt     int64  `json:"end_at"`
// 	TotalTime string `json:"total_time"`
// 	Msg       string `json:"msg"`
// 	Total     int64  `json:"total"`  // 执行总数
// 	Num       int64  `json:"num"`    // 已执行数量
// 	Status    string `json:"status"` // 执行状态
// }

// func (rt *RuningTask) Sent(msg string) {
// 	if dt, err := rt.genRunnerMsg(msg); err == nil {
// 		if bt, err := config.Json.Marshal(map[string]any{
// 			"type": "task",
// 			"data": dt,
// 		}); err == nil {
// 			ws.SentMsg(rt.AdminID, bt)
// 		}
// 	}
// }

// // 获取所有的在运行的任务
// func GetRuningTasks(adminID int64) []*WsTask {
// 	var tasks []*WsTask
// 	WatchingTasks.Range(func(k, v any) bool {
// 		if taskID, ok := k.(int64); ok {
// 			if rt, err := GetRunTask(taskID); err == nil {
// 				if bt, err := rt.genRunnerMsg(""); err == nil {
// 					tasks = append(tasks, bt)
// 				}
// 			}
// 		}

// 		return true
// 	})
// 	return tasks
// }

// // 结构化在运行任务的输出
// func (rt *RuningTask) genRunnerMsg(msg string) (*WsTask, error) {
// 	var rns []map[string]any
// 	rt.mu.Lock()
// 	for _, v := range rt.runners {
// 		rns = append(rns, map[string]any{
// 			"runid": v.s.GetID(),
// 		})
// 	}
// 	rt.mu.Unlock()

// 	return &WsTask{
// 		UUID:   rt.UUID,
// 		ID:     rt.ID,
// 		Msg:    msg,
// 		IsRun:  rt.isRun.Load(),
// 		Title:  rt.Title,
// 		Tags:   rt.Tags,
// 		ErrMsg: rt.ErrMsg,
// 	}, nil
// 	// return map[string]any{
// 	// 	"id":      rt.ID,
// 	// 	"uuid":    rt.UUID,
// 	// 	"isrun":   rt.isRun.Load(),
// 	// 	"length":  len(rns),
// 	// 	"title":   rt.Title,
// 	// 	"tags":    rt.Tags,
// 	// 	"err_msg": rt.ErrMsg,
// 	// 	"runners": rns,
// 	// }
// }
