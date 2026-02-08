package task

type TaskBody struct {
	Code     string
	SubTasks []SubTask
}

func (t *TaskBody) TotalWeight() float64 {
	var w float64
	for _, s := range t.SubTasks {
		w += s.Weight()
	}
	if w <= 0 {
		return 1
	}
	return w
}
