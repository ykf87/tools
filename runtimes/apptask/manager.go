package apptask

import (
	"fmt"
	"time"
)

type Manager struct {
	rt   *runtime
	stop chan struct{}
}

func New(opt Options) *Manager {
	if opt.Store == nil {
		panic("apptask: Store is required")
	}
	if opt.TickInterval <= 0 {
		opt.TickInterval = time.Second
	}
	return &Manager{
		rt:   newRuntime(opt),
		stop: make(chan struct{}),
	}
}

func (m *Manager) BindDevice(deviceId string, d Delivery) {
	m.rt.bindDevice(deviceId, d)
}

func (m *Manager) UnbindDevice(deviceId string) {
	m.rt.unbindDevice(deviceId)
}

func (m *Manager) Start() {
	fmt.Println("----- 启动apptask")
	go func() {
		ticker := time.NewTicker(m.rt.opt.TickInterval)
		for {
			select {
			case <-ticker.C:
				m.rt.tick()
			case <-m.stop:
				return
			}
		}
	}()
}

func (m *Manager) Stop() {
	close(m.stop)
}

func (m *Manager) Report(runId int64, status int, msg string) {
	m.rt.report(runId, status, msg)
}

func (m *Manager) ForceStop(taskId int64, operator string) error {
	return m.rt.forceStop(taskId, operator)
}
