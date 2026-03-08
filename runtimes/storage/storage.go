package storage

import (
	"context"
	"io"
	"net/http"
	"time"
)

type DownloadOption struct {

	// 目标目录
	Dir string

	// 文件名 (可空)
	FileName string

	// HTTP header
	Headers http.Header

	// cookie
	Cookie string

	// 代理
	Proxy string

	// 超时
	Timeout time.Duration

	// 进度
	OnProgress Progress
}

type Storage interface {

	// 上传
	Put(path string, reader io.Reader) error

	// 下载
	Get(path string) (io.ReadCloser, error)

	// 删除
	Delete(path string) error

	// 是否存在
	Exists(path string) (bool, error)

	// 移动
	Move(src, dst string) error

	// 复制
	Copy(src, dst string) error

	// 获取URL
	URL(path string) string

	Download(ctx context.Context, url string, opt DownloadOption) (string, error)
}
