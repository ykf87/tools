// 文件下载
package submqs

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"tools/runtimes/downloader"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"
	"tools/runtimes/mainsignal"
	"tools/runtimes/mq"
)

func init() {
	download()
}

func download() {
	mq.MqClient.Register("download", func(ctx context.Context, msg *mq.Message) error {
		rrs := new(downloader.DownloadOption)
		if err := json.Unmarshal([]byte(msg.Payload), rrs); err != nil {
			logs.Error(err.Error() + ": downloader mq")
			return err
		}

		if rrs.FileName == "" {
			if strings.Contains(rrs.Dir, ".") {
				rrs.FileName = path.Base(rrs.Dir)
				rrs.Dir = path.Dir(rrs.Dir)
			}
		}

		rrs.Callback = func(total, downloaded, speed, workers int64) {
			fmt.Printf(
				"\r%.2f%% %s/s workers:%d %s",
				float64(downloaded)/float64(total)*100,
				funcs.FormatFileSize(speed, "1", ""),
				workers,
				funcs.FormatFileSize(total, "1", ""),
			)
		}
		if _, err := downloader.Download(mainsignal.MainCtx, rrs); err != nil {
			return err
		}

		return nil
		// loader := downloader.NewDownloader(rrs.Proxy, func(percent float64, dlownloaded, total int64) {
		// 	select {
		// 	case <-ctx.Done():
		// 		// worker 停止，安全退出回调
		// 		return
		// 	default:
		// 		eventbus.Bus.Publish("ws", map[string]any{
		// 			"downloaded":    percent,
		// 			"downloadedInt": dlownloaded,
		// 			"total":         total,
		// 		})
		// 	}
		// }, nil)
		// // return loader.Download(rrs.Url, rrs.Dest)
		// // 任务带 ctx 超时 / 取消
		// doneCh := make(chan error, 1)
		// go func() {
		// 	doneCh <- loader.Download(rrs.Url, rrs.Dest)
		// }()

		// select {
		// case <-ctx.Done():
		// 	logs.Warn("download canceled: " + rrs.Url)
		// 	return ctx.Err()
		// case err := <-doneCh:
		// 	if err != nil {
		// 		logs.Error("download failed: " + err.Error())
		// 	}
		// 	return err
		// }
	}, 100)
}
