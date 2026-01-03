package tasks

import (
	"context"
	"sync"
	db "tools/runtimes/db"
	"tools/runtimes/db/clients"
)

func execTask(ctx context.Context, task *Task, runID int64, runData []byte) error {
	var wg sync.WaitGroup

	var sem chan struct{}
	if task.SeNum > 0 {
		sem = make(chan struct{}, task.SeNum)
	}

	var devices []*TaskClients
	db.TaskDB.Where("task_id = ?", task.ID).Find(&devices)

	for _, v := range devices {
		wg.Add(1)

		go func(v *TaskClients) {
			if sem != nil {
				sem <- struct{}{}
			}

			defer func() {
				if sem != nil {
					<-sem
				}
				wg.Done()
			}()

			select {
			case <-ctx.Done():
				// 任务被取消 / 超时
				return
			default:
			}

			// ✅ 真正执行,其实就是执行js代码
			// 至于数据则通过js代码再跟服务端获取
			// 因此
			switch v.DeviceType {
			case 0: // 执行浏览器
				bs, err := clients.GetBrowserById(v.DeviceID)
				if err == nil {
					if err := bs.Open(); err == nil {
						bs.Browser.RunJs("alert('sdfsdf')")
					}
				}
			case 1: // 执行autojs
				phone, err := clients.GetPhoneById(v.DeviceID)
				if err == nil {
					clients.Hubs.SentClient(phone.DeviceId, runData)
				}
			}
			// fmt.Println(v.DeviceID, runID, runData)
		}(v)
	}

	wg.Wait()
	return nil
}
