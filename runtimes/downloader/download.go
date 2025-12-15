package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"tools/runtimes/db/mqs"
	"tools/runtimes/eventbus"
	"tools/runtimes/logs"
	"tools/runtimes/mq"

	"github.com/h2non/filetype"
	jsoniter "github.com/json-iterator/go"
)

// Downloader 配置
type Downloader struct {
	Client     *http.Client                                   `json:"-"`
	OnProgress func(percent float64, downloaded, total int64) `json:"-"`
	Proxy      string                                         `json:"proxy"`
	Url        string                                         `json:"url"`
	Dest       string                                         `json:"dest"`
	Headers    http.Header                                    `json:"headers"`
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	mq.MqClient.Subscribe("download", func(msg *mqs.Mq) error {
		rrs := new(Downloader)
		if err := json.Unmarshal([]byte(msg.Payload), rrs); err != nil {
			logs.Error(err.Error() + ": downloader mq")
			return err
		}
		loader := NewDownloader(rrs.Proxy, func(percent float64, dlownloaded, total int64) {
			eventbus.Bus.Publish("ws", map[string]any{
				"downloaded":    percent,
				"downloadedInt": dlownloaded,
				"total":         total,
			})
		}, nil)
		return loader.Download(rrs.Url, rrs.Dest)
	})
}

// NewDownloader 创建下载器
func NewDownloader(proxy string, onProgress func(percent float64, downloaded, total int64), headers http.Header) *Downloader {
	client := &http.Client{}

	// 设置代理
	if proxy != "" {
		if proxyURL, err := url.Parse(proxy); err == nil {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		}
	}

	return &Downloader{
		Client:     client,
		Proxy:      proxy,
		OnProgress: onProgress,
		Headers:    headers,
	}
}

// Download 下载文件，支持断点续传. destPath需要完整的文件名称
func (d *Downloader) Download(urlStr, destPath string) error {
	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}

	// 检查本地文件大小
	var startPos int64 = 0
	var file *os.File
	if info, err := os.Stat(destPath); err == nil {
		startPos = info.Size()
		file, err = os.OpenFile(destPath, os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		file, err = os.Create(destPath)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	// 创建请求
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}

	// 设置 Range 头，支持断点续传
	if startPos > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startPos))
	}

	if d.Headers != nil {
		req.Header = d.Headers
	}

	// 发起请求
	resp, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 获取总大小
	var totalSize int64
	if resp.Header.Get("Content-Range") != "" {
		var end int64
		_, err := fmt.Sscanf(resp.Header.Get("Content-Range"), "bytes %d-%d/%d", &startPos, &end, &totalSize)
		if err != nil {
			return fmt.Errorf("解析 Content-Range 出错: %v", err)
		}
	} else {
		totalSize = resp.ContentLength + startPos
	}

	// 写入数据并更新进度
	buf := make([]byte, 32*1024)
	var downloaded = startPos
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			// 回调进度
			if totalSize > 0 && d.OnProgress != nil {
				percent := float64(downloaded) / float64(totalSize) * 100
				d.OnProgress(percent, downloaded, totalSize)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

func Down(url, dest, proxy string) {
	mq.MqClient.Publish("download", Downloader{Proxy: proxy, Url: url, Dest: dest}, 0)
}

func (d *Downloader) GetUrlFileExt(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := d.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	headerBuf := make([]byte, 261)
	n, err := resp.Body.Read(headerBuf)
	if err != nil && err != io.EOF {
		return "", err
	}

	kind, _ := filetype.Match(headerBuf[:n])
	var ext string
	if kind != filetype.Unknown {
		ext = kind.Extension
	} else {
		// fallback: 尝试从 URL 获取扩展名
		ext = strings.TrimLeft(filepath.Ext(url), ".")
	}
	return ext, nil
}

// func main() {
// 	url := "https://speed.hetzner.de/100MB.bin"
// 	dest := "downloads/test.bin"
// 	proxy := "http://127.0.0.1:7890" // 如果不需要代理，设为空 ""

// 	d := downloader.NewDownloader(proxy, func(percent float64) {
// 		fmt.Printf("\r下载进度: %.2f%%", percent)
// 	})

// 	if err := d.Download(url, dest); err != nil {
// 		fmt.Println("\n下载失败:", err)
// 	} else {
// 		fmt.Println("\n下载完成!")
// 	}
// }
