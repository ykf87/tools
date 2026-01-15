package bs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db/messages"
	"tools/runtimes/db/notifies"
	"tools/runtimes/funcs"
	"tools/runtimes/services"
)

func DownBrowserBinFile(saveto string) error {
	nty := notifies.NewNotify()
	nty.Type = "start"
	nty.Title = "指纹浏览器 - 可在任务列表查看"
	nty.Meta = time.Now().Format("2006-01-02 15:04:05")
	nty.Description = "您的操作依赖指纹浏览器,正在下载..."
	nty.KeepOpen = true
	nty.Content = `什么是指纹浏览器:指纹浏览器是一种用于浏览器指纹识别的技术工具，采用修改浏览器指纹的方法来掩盖用户的真实身份和设备信息。
	指纹浏览器通过利用模拟浏览器硬件配置文件来实现浏览器指纹防护功能。
	用户下载安装特殊的浏览器以使用指纹浏览器，通过自定义或修改浏览器指纹以改变配置信息，使其看起来来自不同的设备和地点。`
	go nty.Send()

	var serverBrowserName string
	switch runtime.GOOS {
	case "darwin":
		serverBrowserName = "browser-mac.zip"
	default:
		serverBrowserName = "browser-win.zip"
	}
	downurl := fmt.Sprint(config.SERVERDOMAIN, "down?file=browser/"+serverBrowserName)
	saveFile := filepath.Join(saveto, serverBrowserName)

	if err := services.ServerDownload(downurl, saveFile, nil, func(perc float64, downloaded, total int64) {
		nty.Schedule = perc
		go nty.Send()
		fmt.Printf("\r下载中：%.2f%% (%s/%s)", perc,
			funcs.FormatFileSize(downloaded),
			funcs.FormatFileSize(total))
	}); err != nil {
		nty.Description = err.Error()
		nty.Url = config.ApiUrl + "/browser/download"
		nty.Btn = "重新下载"
		nty.Method = "get"
		nty.KeepOpen = false
		nty.Done = true
		nty.Send()
		return err
	}
	nty.Description = "下载完成,开始解压......"
	nty.Schedule = 100
	go nty.Send()

	if err := funcs.Unzip(saveFile, saveto); err != nil {
		nty.Meta = time.Now().Format("2006-01-02 15:04:05")
		nty.Description = err.Error()
		nty.Url = config.ApiUrl + "/browser/download"
		nty.Btn = "重新下载"
		nty.Method = "get"
		nty.KeepOpen = false
		nty.Done = true
		go nty.Send()
		return err
	}

	_ = os.Remove(saveFile)

	msg := &messages.Message{
		Type:    "success",
		Content: "指纹浏览器下载成功!",
	}
	go msg.Send()

	return nil
}
