scheduler := NewWithLimit(2)

### 1. 任务体（定义流程）
body := &task.TaskBody{
	Code: "sync_user",
	SubTasks: []task.SubTask{
		&DownloadSubTask{},
		&ProcessSubTask{},
	},
}

### 2. 落库任务实例
taskIns := &task.TaskInstance{
	ID:       1001,
	TaskCode: "sync_user",
	Status:   task.StatusPending,
}

### 3. Executor（Runner 调用的）
executor := &task.TaskExecutor{
	Body:     body,
	Instance: taskIns,
	Reporter: task.NewLogReporter(),
	CreateSub: func(taskID int64, name string) *task.SubTaskInstance {
		// 这里你落库，返回 subtask 实例
		return &task.SubTaskInstance{
			ID:     genID(),
			TaskID: taskID,
			Name:   name,
			Status: task.StatusPending,
		}
	},
}

### 4. Runner
runner := scheduler.NewRunner(executor.Execute)
runner.SetMaxTry(3)
runner.SetError(func(err error) {
	// runner 级错误
})
runner.Run()
