package tasks

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/proxys"
	"tools/runtimes/db/task"
	"tools/runtimes/expire"
	"tools/runtimes/i18n"
	"tools/runtimes/mainsignal"
	"tools/runtimes/proxy"
	"tools/runtimes/runner"
)

var Runners sync.Map

func (t *Task) ReStart() {
	t.Stop()
	time.Sleep(time.Millisecond * 100)
	t.Run()
}

// 启动任务
func (t *Task) Run() error {
	if sc := getRunner(t.ID); sc != nil {
		return nil
	}

	return t.build()
}

// 停止任务
func (t *Task) Stop() {
	if sc := getRunner(t.ID); sc != nil {
		sc.StopAll()
	}
	Runners.Delete(t.ID)
}

// 清空
func Flush() {

}

func getRunner(id int64) *task.Task {
	if ttr, ok := RunnerTasks.Load(id); ok {
		if scr, ok := ttr.(*task.Task); ok {
			return scr
		}
	}
	return nil
}

// 构建任务启动
func (t *Task) build() error {
	if t.Endtime > 0 && t.Endtime < time.Now().Unix() {
		return errors.New(i18n.T("Expired"))
	}

	// 查找执行客户端
	if len(t.Clients) < 1 {
		t.Clients = t.GetClients()
	}
	if len(t.Clients) < 1 {
		return errors.New(i18n.T("No Client to run"))
	}

	if t.SeNum < 1 {
		t.SeNum = 2
	}

	sc, err := task.NewTask("task", t.ID, t.Title, t.SeNum, false)
	if err != nil {
		return err
	}

	for _, v := range t.Clients {
		var name string
		switch v.DeviceType {
		case 1:
			name = fmt.Sprintf("手机端: %d", v.DeviceID)
		case 2:
			name = fmt.Sprintf("HTTP: %d", v.DeviceID)
		default:
			name = fmt.Sprintf("Web端: %d", v.DeviceID)
		}
		tsc, err := sc.AddInterval(
			t.genRunnerId(v),
			fmt.Sprintf("%s", name),
			time.Duration(t.Cycle*60)*time.Second,
			time.Duration(t.Timeout)*time.Second,
			t.RetryMax,
			time.Second*20,
			time.Unix(t.Endtime, 0),
			func(tr *task.TaskRun) error {
				return t.callback(v)
			},
		)
		if err != nil {
			return err
		}
		if t.Starttime > 0 {
			tsc.SetStartAt(time.Unix(t.Starttime, 0))
			tsc.RunNow()
		}
	}
	RunnerTasks.Store(t.ID, sc)

	if t.Endtime > 0 {
		expire.Add(t.Endtime, func() {
			t.Stop()
		})
	}

	return nil
}

// 生成子任务id
func (t *Task) genRunnerId(v *TaskClients) string {
	return fmt.Sprintf("%d:%d:%d", t.ID, v.DeviceType, v.DeviceID)
}

// 任务回调
func (t *Task) callback(v *TaskClients) error {
	r, err := t.genrunner(v)
	if err != nil {
		return err
	}
	return r.Start(time.Duration(t.Timeout), func(s string) error {
		return nil
	})
}

// 生成runner的配置
func (t *Task) genrunner(v *TaskClients) (runner.Runner, error) {
	var opt any

	switch v.DeviceType {
	case 1:
		opt = runner.GenPhoneOpt()
	case 2:
		opt = runner.GenHttpOpt()
	default:
		bs, err := browserdb.GetBrowserById(v.DeviceID)
		if err != nil {
			return nil, err
		}
		var pc *proxy.ProxyConfig
		if bs.Proxy > 0 {
			if px := proxys.GetById(bs.Proxy); px != nil {
				if pcc, err := proxy.Client(px.GetConfig(), "", 0, px.GetTransfer()); err == nil {
					pc = pcc
				}
			} else if bs.ProxyConfig != "" {
				if pcc, err := proxy.Client(bs.ProxyConfig, "", 0); err == nil {
					pc = pcc
				}
			}
		}
		if t.Timeout < 1 {
			t.Timeout = 60
		}
		opt = runner.GenWebOpt(
			mainsignal.MainCtx,
			v.DeviceID,
			t.Headless != 1,
			t.DefUrl, t.GetRunJscript(),
			pc,
			time.Duration(t.Timeout),
			bs.Width,
			bs.Height,
			bs.Lang,
			bs.Timezone,
		)
	}

	return runner.GetRunner(v.DeviceType, opt)
}
