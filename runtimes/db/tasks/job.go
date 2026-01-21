package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tools/runtimes/scheduler"
)

var tmpIndex int64
var dbTasks sync.Map

func TempTask(adminID int64) *Task {
	tmpIndex--
	return &Task{
		ID:     tmpIndex,
		Status: 1,
	}
}

func listen() {
	for {
		dbTasks.Range(func(k, v any) bool {
			if tsk, ok := v.(*Task); ok {
				tsk.watching()
			}
			return true
		})
		time.Sleep(time.Second * 5)
	}
}

func (this *Task) watching() error {
	if this.ID == 0 {
		return fmt.Errorf("任务未设置")
	}

	now := time.Now().Unix()
	if this.Starttime > 0 && this.Starttime > now {
		return nil
	}

	if this.Endtime > 0 && this.Endtime < now {
		return nil
	}
	this.push()

	return nil
}

// 将任务添加到调度器并启动
func (this *Task) push() error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.ID == 0 {
		return fmt.Errorf("任务未设置")
	}

	taskID := this.taskID()

	if Seched.Exists(taskID) {
		return nil
	}

	go func() {
		time.Sleep(time.Second * 14)
		Seched.Remove(this.taskID())
		fmt.Println("发送删除任务---")
	}()

	return Seched.Add(&scheduler.Task{
		ID:       taskID,
		Interval: time.Duration(this.Cycle) * time.Second,
		Run:      this.runner,
		Stop:     this.stop,
		MaxRetry: this.RetryMax,
		Timeout:  time.Duration(this.Timeout) * time.Second,
	})

}

func (this *Task) runner(ctx context.Context) error {
	fmt.Println(this.ID, "-----", this.RetryMax)
	time.Sleep(time.Second * 8)
	return nil //fmt.Errorf("error")
}

func (this *Task) stop(ctx context.Context) error {
	// this.mu.Lock()
	// defer this.mu.Unlock()
	fmt.Println("开始停止任务了---")

	// Seched.Remove(this.taskID())
	this.isRuning = false
	if tt, ok := dbTasks.Load(this.ID); ok {
		if tsk, ok := tt.(*Task); ok {
			tsk.Status = 0
			tsk.Save(nil)
		}
		dbTasks.Delete(this.ID)
	}
	fmt.Println("任务清理")
	return nil
}

func (this *Task) taskID() string {
	return fmt.Sprintf("task-%d", this.ID)
}
