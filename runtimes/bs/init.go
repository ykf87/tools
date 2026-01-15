package bs

import (
	"os"
	"path/filepath"
	"runtime"
	"tools/runtimes/config"
)

func init() {
	BROWSERPATH = filepath.Join(config.SYSROOT, "browser")
}

func getBrowserBinName() (string, error) {
	var filename string
	switch runtime.GOOS {
	case "darwin":
		filename = "VirtualBrowser.dmg"
	default:
		filename = "VirtualBrowser.exe"
	}
	// 如果不存在执行文件,则下载
	fullName := filepath.Join(BROWSERPATH, filename)
	if _, err := os.Stat(fullName); err != nil {
		DownBrowserBinFile(BROWSERPATH)
	}
	return config.FullPath(fullName), nil
}
