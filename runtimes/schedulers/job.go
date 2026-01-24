package schedulers

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/go-co-op/gocron"
)

type ExecFun func(ctx context.Context, job *Job) error

type Job struct {
	id      string
	ctx     context.Context
	cancel  context.CancelFunc
	fn      ExecFun
	maxTry  int
	tried   int
	s       *gocron.Scheduler
	jobHand *gocron.Job
	mu      sync.Mutex
	isrun   atomic.Bool
}

func (j *Job) run() error {
	select {
	case <-j.ctx.Done():
		j.isrun.Store(false)
		return j.ctx.Err()
	default:
	}

	defer j.isrun.Store(false)

	err := j.fn(j.ctx, j)
	if err != nil {
		j.tried++
		if j.tried >= j.maxTry {
			j.cancel()
		}
		return err
	}

	j.tried = 0
	return nil
}

func (j *Job) Stop() {
	fmt.Println("停止任务---")
	j.s.Remove(j.jobHand)
	j.cancel()
}

func (j *Job) GetName() string {
	return j.jobHand.GetName()
}
func (j *Job) GetID() string {
	return j.id
}
