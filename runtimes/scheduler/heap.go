package scheduler

import "time"

type runnerHeap []*Runner

func (h runnerHeap) Len() int { return len(h) }

func (h runnerHeap) Less(i, j int) bool {
	return h[i].nextRun.Before(h[j].nextRun)
}

func (h runnerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *runnerHeap) Push(x any) {
	*h = append(*h, x.(*Runner))
}

func (h *runnerHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

func (h runnerHeap) Peek() *Runner {
	if len(h) == 0 {
		return nil
	}
	return h[0]
}

func (h runnerHeap) NextDuration(now time.Time) time.Duration {
	if len(h) == 0 {
		return -1
	}
	return h[0].nextRun.Sub(now)
}
