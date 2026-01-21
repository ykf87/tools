package scheduler

type taskHeap []*Task

func (h taskHeap) Len() int { return len(h) }

func (h taskHeap) Less(i, j int) bool {
	return h[i].NextRun.Before(h[j].NextRun)
}

func (h taskHeap) Swap(i, j int) {
	// ✅ 防御式检查（生产级必须）
	if i < 0 || j < 0 || i >= len(h) || j >= len(h) {
		return
	}
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *taskHeap) Push(x any) {
	t := x.(*Task)
	t.index = len(*h)
	*h = append(*h, t)
}

func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	if n == 0 {
		return nil
	}
	t := old[n-1]
	t.index = -1 // ⭐⭐ 关键：不在 heap 了
	*h = old[:n-1]
	return t
}
