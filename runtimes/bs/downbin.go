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

func serverName() string {
	return runtime.GOOS + "/browser.zip"
}

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

	serverBrowserName := serverName()
	// switch runtime.GOOS {
	// case "darwin":
	// 	serverBrowserName = "browser-mac.zip"
	// default:
	// 	serverBrowserName = "browser-win.zip"
	// }
	// downurl := fmt.Sprint(config.SERVERDOMAIN, "down?file=browser/"+serverBrowserName)
	saveFile := filepath.Join(saveto, serverBrowserName)

	if err := services.ServerDownload(serverBrowserName, filepath.Dir(saveFile), nil, func(total, downloaded, speed, workers int64) {
		msgstr := fmt.Sprintf(
			"%.2f%% %s/s %s 线程: %d",
			float64(downloaded)/float64(total)*100,
			funcs.FormatFileSize(speed, "1", ""),
			funcs.FormatFileSize(total, "1", ""),
			workers,
		)
		fmt.Print("\r", msgstr)
	}); err != nil {
		nty.Description = err.Error()
		nty.Url = config.ApiUrl + "/browser/download"
		nty.Btn = "重新下载"
		nty.Method = "get"
		nty.KeepOpen = false
		nty.Done = true
		nty.Send()
		fmt.Println("浏览器下载失败")
		return err
	}
	nty.Description = "下载完成,开始解压 浏览器......"
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
