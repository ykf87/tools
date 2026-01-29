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

// func (t *TaskClients) start() {
// 	go func() {
// 		if err := t.acquire(); err != nil {
// 			return
// 		}
// 		var runner Runner
// 		switch t.tsk.Tp {
// 		case 0:
// 			runner = runweb.New(t.release, &runweb.Option{
// 				Headless: !(t.tsk.Headless == 1),
// 				Js:       t.tsk.GetRunJscript(),
// 				Url:      t.tsk.DefUrl,
// 				ID:       t.DeviceID,
// 				Timeout:  time.Duration(t.tsk.Timeout) * time.Second,
// 				OnError:  t.tsk.OnError,
// 				OnClose:  t.tsk.OnClose,
// 				OnChange: t.tsk.OnChange,
// 				Callback: t.tsk.Callback,
// 			})
// 		case 1:
// 			runner = runphone.New(t.release)
// 		case 2:
// 			runner = runhttp.New(t.release)
// 		default:
// 			t.tsk.ErrMsg = "类型不被支持"
// 			go t.tsk.Sent()
// 			return
// 		}

// 		t.runner = Seched.
// 			NewRunner(runner.Start, time.Duration(t.tsk.Timeout)*time.Second).
// 			Every(time.Duration(t.tsk.Cycle) * time.Second).
// 			SetCloser(runner.OnClose).
// 			SetError(runner.OnError).
// 			SetMaxTry(t.tsk.RetryMax).
// 			SetRetryDelay(time.Second * 10)
// 		runner.SetRunner(t.runner)
// 		t.runner.RunNow()
// 		fmt.Println(t.tsk.SeNum, t.tsk.Cycle, "---- 开启了taskclients", t.DeviceType, t.DeviceID, t.tsk.ID)
// 	}()
// }

// 阻塞式：严格控制并发
// func (t *TaskClients) acquire() error {
// 	if t.tsk == nil || t.tsk.slots == nil {
// 		return errors.New("task 未初始化")
// 	}

// 	t.tsk.slots <- struct{}{}
// 	return nil
// 	// select {
// 	// case t.tsk.slots <- struct{}{}:
// 	// 	fmt.Println("写入----")
// 	// 	return nil
// 	// default:
// 	// 	// return t.tsk.ctx.Err()
// 	// 	return nil
// 	// }
// }

// // release
// func (t *TaskClients) release() {
// 	select {
// 	case <-t.tsk.slots:
// 	default:
// 		// 理论上不应该发生，防御式
// 	}
// }

// 非阻塞尝试获取
// func (t *TaskClients) tryAcquire() bool {
// 	select {
// 	case t.tsk.slots <- struct{}{}:
// 		return true
// 	default:
// 		return false
// 	}
// }

// // 超时控制
// func (t *TaskClients) acquireWithTimeout(d time.Duration) error {
// 	select {
// 	case t.tsk.slots <- struct{}{}:
// 		return nil
// 	case <-time.After(d):
// 		return errors.New("acquire timeout")
// 	}
// }

func (t *TaskClients) GetName() string {
	return fmt.Sprintf("%d%d%d", t.TaskID, t.DeviceType, t.DeviceID)
}
