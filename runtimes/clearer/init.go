package clearer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"
	"tools/runtimes/services"
)

const (
	SERVERNAME = "clearer.zip"
	LOCALFILE  = "clearer.exe"
)

var BaseRoot = config.FullPath(config.SYSROOT, "clearer")
var FullFileName = filepath.Join(BaseRoot, LOCALFILE)

func Init() {
	if _, err := os.Stat(BaseRoot); err != nil {
		download(BaseRoot)
	}
}

func download(ddir string) error {
	fmt.Println("下载 图片高清工具")

	if err := services.ServerDownload(serverName(), ddir, nil, func(total, downloaded, speed, workers int64) {
		msgstr := fmt.Sprintf(
			"%.2f%% %s/s %s 线程: %d",
			float64(downloaded)/float64(total)*100,
			funcs.FormatFileSize(speed, "1", ""),
			funcs.FormatFileSize(total, "1", ""),
			workers,
		)
		fmt.Print("\r", msgstr)
	}); err != nil {
		fmt.Println("图片高清工具 下载失败", err, serverName())
		return err
	}

	fmt.Println("\n下载完成,开始解压......")
	if err := unzip(config.FullPath(ddir, SERVERNAME), ddir); err != nil {
		return err
	}

	config.FFmpeg = ddir
	return nil
}
func unzip(fullname, zipto string) error {
	if err := funcs.Unzip(fullname, zipto); err != nil {
		logs.Error("解压 图片高清工具 失败:" + err.Error())
		return err
	}
	os.Remove(fullname)

	fmt.Println("解压完成!")
	return nil
}

func serverName() string {
	return runtime.GOOS + "/" + SERVERNAME
}
