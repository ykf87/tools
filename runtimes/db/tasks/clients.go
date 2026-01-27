package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"
	"tools/runtimes/scheduler"
)

// 任务设备表
type TaskClients struct {
	TaskID     int64             `json:"task_id" gorm:"primaryKey"`     // 任务ID
	DeviceType int               `json:"device_type" gorm:"primaryKey"` // 设备类型, 0-web端 1-手机端 2-发起http请求
	DeviceID   int64             `json:"device_id" gorm:"primaryKey"`   // 设备id,在clients下的 browsers 或者 phones 表
	Url        string            `json:"url" gorm:"primaryKey"`         // http发起的url, 从此处开始为 device_type = 2的参数
	Method     string            `json:"method"`                        // http 请求方式
	Data       string            `json:"data"`                          // 如果是post发起的，携带数据
	Cookies    string            `json:"cookies"`                       // 使用的cookie
	Headers    string            `json:"headers"`                       // 携带的头部信息
	tsk        *Task             `json:"-" gorm:"-"`                    // 任务
	runner     *scheduler.Runner `json:"-" gorm:"-"`                    // 调度器中的任务
}

func (t *TaskClients) start() {
	go func() {
		if err := t.acquire(); err != nil {
			return
		}
		t.runner = Seched.NewRunner(t.run)
		if t.tsk.Cycle > 0 {
			t.runner.Every(time.Second * time.Duration(t.tsk.Cycle))
		}
		if t.tsk.RetryMax > 0 {
			t.runner.SetMaxTry(t.tsk.RetryMax)
		}

		if t.tsk.OnClose != nil {
			t.runner.SetCloser(t.tsk.OnClose)
		}
		if t.tsk.OnError != nil {
			t.runner.SetError(t.tsk.OnError)
		}
		t.runner.SetRetryDelay(time.Second * 10)

		t.runner.RunNow()
		fmt.Println("---- 开启了taskclients", t.DeviceType, t.DeviceID, t.tsk.ID)
	}()
}

// 阻塞式：严格控制并发
func (t *TaskClients) acquire() error {
	if t.tsk == nil || t.tsk.slots == nil {
		return errors.New("task 未初始化")
	}

	select {
	case t.tsk.slots <- struct{}{}:
		return nil
	case <-t.tsk.ctx.Done():
		return t.tsk.ctx.Err()
	}
}

// release
func (t *TaskClients) release() {
	select {
	case <-t.tsk.slots:
	default:
		// 理论上不应该发生，防御式
	}
}

// 非阻塞尝试获取
func (t *TaskClients) tryAcquire() bool {
	select {
	case t.tsk.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

// 超时控制
func (t *TaskClients) acquireWithTimeout(d time.Duration) error {
	select {
	case t.tsk.slots <- struct{}{}:
		return nil
	case <-time.After(d):
		return errors.New("acquire timeout")
	}
}

// 执行入口
func (t *TaskClients) run(ctx context.Context) error {
	defer t.release()

	switch t.tsk.Tp {
	case 0:
		return t.runBrowser()
	case 1:
		return t.runPhone()
	case 2:
		return t.runHttp()
	default:
		return fmt.Errorf("不支持的类型")
	}
}

// 执行浏览器
func (t *TaskClients) runBrowser() error {
	time.Sleep(time.Second * 6)
	return nil
}

// 执行手机
func (t *TaskClients) runPhone() error {
	return nil
}

// 执行http
func (t *TaskClients) runHttp() error {
	return nil
}
