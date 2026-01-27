package tasks

import "tools/runtimes/eventbus"

func (t *Task) Sent() {
	eventbus.Bus.Publish("task", t)
}
