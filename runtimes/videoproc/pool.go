package videoproc

import (
	"context"
	"sync"
)

type Task func(ctx context.Context) error

func RunWithCancel(ctx context.Context, concurrency int, tasks []Task) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	taskCh := make(chan Task)
	errCh := make(chan error, 1)

	// workers
	for range concurrency {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-taskCh:
					if !ok {
						return
					}
					if err := t(ctx); err != nil {
						select {
						case errCh <- err:
						default:
						}
						cancel()
						return
					}
				}
			}
		})
	}

	// producer
	go func() {
		defer close(taskCh)
		for _, t := range tasks {
			select {
			case <-ctx.Done():
				return
			case taskCh <- t:
			}
		}
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
