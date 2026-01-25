# scheduler

一个轻量级、无 ticker 的 Go 任务调度器，基于最小 nextRun 调度。

## 特性

- 不使用 time.Ticker
- 最小 nextRun 唤醒
- 默认任务只执行一次
- 显式设置 interval 才会周期执行
- 支持立即执行（RunNow）
- 线程安全

## 安装

```bash
go get github.com/yourname/scheduler


---

## 5️⃣ example_test.go（行为验证）

```go
package scheduler_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yourname/scheduler"
)

func TestRunOnce(t *testing.T) {
	s := scheduler.New()

	s.NewRunner(func(ctx context.Context) error {
		fmt.Println("only once")
		return nil
	}).RunNow()

	time.Sleep(time.Second)
	s.Stop()
}
