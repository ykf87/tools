package scheduler

import (
	"context"
	"fmt"
	"time"
)

func Demo() {
	// ===== 创建调度器 =====
	s := New(Options{
		MaxConcurrency: 2,
		MaxQueueSize:   100,
		PersistFile:    "tasks.json",
	})

	s.Start()
	defer s.Stop()

	// ===== 任务 1：成功任务 =====
	s.Add(&Task{
		ID:       "task-success",
		Interval: 3 * time.Second,
		Timeout:  5 * time.Second,
		Run: func(ctx context.Context) error {
			fmt.Println("[task-success] running")
			time.Sleep(2 * time.Second)
			fmt.Println("[task-success] done")
			return nil
		},
	})

	// ===== 任务 2：失败 + 重试 =====
	failOnce := true
	s.Add(&Task{
		ID:       "task-fail-then-retry",
		Interval: 0,
		Timeout:  3 * time.Second,
		MaxRetry: 2,
		RetryGap: 1 * time.Second,
		Run: func(ctx context.Context) error {
			fmt.Println("[task-fail] running")
			if failOnce {
				failOnce = false
				fmt.Println("[task-fail] fail once")
				return fmt.Errorf("intentional failure")
			}
			fmt.Println("[task-fail] success after retry")
			return nil
		},
	})

	// ===== 任务 3：超时任务（验证互斥） =====
	s.Add(&Task{
		ID:       "task-timeout",
		Interval: 2 * time.Second,
		Timeout:  2 * time.Second,
		Run: func(ctx context.Context) error {
			fmt.Println("[task-timeout] start")
			select {
			case <-time.After(5 * time.Second):
				fmt.Println("[task-timeout] done")
				return nil
			case <-ctx.Done():
				fmt.Println("[task-timeout] timeout")
				return ctx.Err()
			}
		},
	})

	// ===== 定期观察调度器状态 =====
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 5; i++ {
		<-ticker.C

		fmt.Println("\n===== Scheduler State =====")

		fmt.Println("Running Tasks:")
		for _, t := range s.RunningTasks() {
			fmt.Printf(" - %s (since %s)\n", t.ID, t.LastRun.Format(time.RFC3339))
		}

		fmt.Println("Waiting Tasks:")
		for _, t := range s.WaitingTasks() {
			fmt.Printf(" - %s (next %s)\n", t.ID, t.NextRun.Format(time.RFC3339))
		}

		fmt.Println("Execution Logs:")
		for _, log := range s.Logs() {
			fmt.Printf(
				" - [%s] success=%v duration=%v retry=%d err=%s\n",
				log.TaskID,
				log.Success,
				log.Duration,
				log.RetryCount,
				log.Error,
			)
		}
	}
}
