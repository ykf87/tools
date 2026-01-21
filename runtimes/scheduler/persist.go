package scheduler

import (
	"container/heap"
	"encoding/json"
	"os"
)

func (s *Scheduler) save() {
	if s.opts.PersistFile == "" {
		return
	}

	data, _ := json.Marshal(s.tasks)
	_ = os.WriteFile(s.opts.PersistFile, data, 0644)
}

func (s *Scheduler) load() {
	if s.opts.PersistFile == "" {
		return
	}
	data, err := os.ReadFile(s.opts.PersistFile)
	if err != nil {
		return
	}

	var m map[string]*Task
	if json.Unmarshal(data, &m) == nil {
		for _, t := range m {
			s.tasks[t.ID] = t
			heap.Push(&s.heap, t)
		}
	}
}
