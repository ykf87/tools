// 文件下载
package submqs

import (
	"context"
	"encoding/json"
	"tools/runtimes/downloader"
	"tools/runtimes/eventbus"
	"tools/runtimes/logs"
	"tools/runtimes/mq"
)

func init() {
	download()
}

func download() {
	mq.MqClient.Register("download", func(ctx context.Context, msg *mq.Message) error {
		rrs := new(downloader.Downloader)
		if err := json.Unmarshal([]byte(msg.Payload), rrs); err != nil {
			logs.Error(err.Error() + ": downloader mq")
			return err
		}
		loader := downloader.NewDownloader(rrs.Proxy, func(percent float64, dlownloaded, total int64) {
			select {
			case <-ctx.Done():
				// worker 停止，安全退出回调
				return
			default:
				eventbus.Bus.Publish("ws", map[string]any{
					"downloaded":    percent,
					"downloadedInt": dlownloaded,
					"total":         total,
				})
			}
		}, nil)
		// return loader.Download(rrs.Url, rrs.Dest)
		// 任务带 ctx 超时 / 取消
		doneCh := make(chan error, 1)
		go func() {
			doneCh <- loader.Download(rrs.Url, rrs.Dest)
		}()

		select {
		case <-ctx.Done():
			logs.Warn("download canceled: " + rrs.Url)
			return ctx.Err()
		case err := <-doneCh:
			if err != nil {
				logs.Error("download failed: " + err.Error())
			}
			return err
		}
	}, 100)
}
