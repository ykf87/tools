package tasks

import "tools/runtimes/eventbus"

func (t *Task) Sent() {
	eventbus.Bus.Publish("task", t)
}

// 获取所有的在运行的任务
func GetRuningTasks(adminID int64) []map[string]any {
	var tasks []map[string]any
	WatchingTasks.Range(func(k, v any) bool {
		if taskID, ok := k.(int64); ok {
			if rt, err := GetRunTask(taskID); err == nil {
				tasks = append(tasks, rt.genRunnerMsg())
			}
		}

		return true
	})
	return tasks
}

// 结构化在运行任务的输出
func (rt *RuningTask) genRunnerMsg() map[string]any {
	var rns []map[string]any
	rt.mu.Lock()
	for _, v := range rt.runners {
		rns = append(rns, map[string]any{
			"runid": v.s.GetID(),
		})
	}
	rt.mu.Unlock()
	return map[string]any{
		"id":      rt.ID,
		"uuid":    rt.UUID,
		"isrun":   rt.isRun.Load(),
		"length":  len(rns),
		"title":   rt.Title,
		"tags":    rt.Tags,
		"err_msg": rt.ErrMsg,
		"runners": rns,
	}
}
