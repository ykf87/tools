package tasklog

import (
	"context"
	"tools/runtimes/bs"
	"tools/runtimes/db/clients"
	"tools/runtimes/scheduler"
)

// 具体执行的任务
type TaskRunner struct {
	ID       string                          `json:"id"`    // 调用方设置的id
	Title    string                          `json:"title"` // 执行的标题
	Msg      chan string                     `json:"-"`     // 消息接收器
	ErrMsg   chan string                     `json:"-"`     // 错误消息接收器
	Total    float64                         `json:"total"` // 执行总量,比如下载文件总大小,或者执行养号总数
	Doned    float64                         `json:"doned"` // 已执行次数或已完成数量
	Runner   *scheduler.Runner               `json:"-"`     // 执行器
	taskID   string                          `json:"-"`     // 任务id
	startAt  int64                           `json:"-"`     // 开始执行时间
	endAt    int64                           `json:"-"`     // 结束事件
	callBack func(string, *TaskRunner) error `json:"-"`     // 内容回调
	opt      any                             `json:"-"`     // 配置
	sch      *scheduler.Scheduler            `json:"-"`     // 调度器,和Task调度器一致
	bss      *bs.Browser                     `json:"-"`     // 浏览器
	phone    clients.Phone                   `json:"-"`     // 手机端
	ctx      context.Context                 `json:"-"`     // 上下文
	cancle   context.CancelFunc              `json:"-"`     // 关闭
	// callback scheduler.TaskFunc   `json:"-"`        // 执行回调
	// onerror  scheduler.ErrFun     `json:"-"`        // 错误回调
	// onclose  scheduler.CloseFun   `json:"-"`        // 结束回调
}

// 关闭执行
func (tr *TaskRunner) Close() {
	tr.cancle()
}
