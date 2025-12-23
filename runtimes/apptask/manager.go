package apptask

import "time"

type Manager struct {
	rt   *runtimeManager
	stop chan struct{}
}

func New(opt Options) *Manager {
	if opt.TickInterval <= 0 {
		opt.TickInterval = time.Second
	}
	return &Manager{
		rt:   newRuntime(opt),
		stop: make(chan struct{}),
	}
}

func (m *Manager) Start() {
	go func() {
		ticker := time.NewTicker(m.rt.opt.TickInterval)
		defer ticker.Stop()

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

func (m *Manager) AddTask(task *AppTask) {
	m.rt.add(task)
}
