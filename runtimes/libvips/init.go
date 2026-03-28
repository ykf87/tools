package libvips

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
	VIPSERVERNAME = "vips.zip"
)

var VIPROOT = config.FullPath(config.SYSROOT, "vips")
var VIPBIN = config.FullPath(VIPROOT, "bin")

func Init() {
	if _, err := os.Stat(VIPBIN); err != nil {
		downloadvips(VIPROOT)
	}

	// 	pkgconfigDir := config.FullPath(config.SYSROOT, "pkgconfig")
	// 	if _, err := os.Stat(pkgconfigDir); err != nil {
	// 		os.MkdirAll(pkgconfigDir, os.ModePerm)
	// 	}

	// 	vipspc := fmt.Sprintf(`prefix=%s
	// exec_prefix=${prefix}
	// libdir=${prefix}/lib
	// includedir=${prefix}/include

	// Name: vips
	// Description: VIPS image processing library
	// Version: 8.14.0

	// Libs: -L${libdir} -lvips
	// Cflags: -I${includedir}`, VIPROOT)
	// 	err := os.WriteFile(filepath.Join(pkgconfigDir, "vips.pc"), []byte(vipspc), 0644)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	os.Setenv("PKG_CONFIG", config.FullPath(config.SYSROOT, "pkg-config.exe"))
	// 	os.Setenv("PKG_CONFIG_PATH", pkgconfigDir)

	// kernel32 := syscall.NewLazyDLL("kernel32.dll")
	// setDllDir := kernel32.NewProc("SetDllDirectoryW")

	// dir, _ := syscall.UTF16PtrFromString(VIPBIN)
	// setDllDir.Call(uintptr(unsafe.Pointer(dir)))
}

func downloadvips(ddir string) error {
	fmt.Println("下载 图片处理工具")

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
		fmt.Println("图片处理工具 下载失败", err, serverName())
		return err
	}

	fmt.Println("\n下载完成,开始解压......")
	if err := unzip(config.FullPath(ddir, VIPSERVERNAME), ddir); err != nil {
		return err
	}

	config.FFmpeg = ddir
	return nil
}

func unzip(fullname, zipto string) error {
	if err := funcs.Unzip(fullname, zipto); err != nil {
		logs.Error("解压 图片处理工具 失败:" + err.Error())
		return err
	}
	os.Remove(fullname)

	fmt.Println("解压完成!")
	return nil
}

func serverName() string {
	return runtime.GOOS + "/" + VIPSERVERNAME
}

func Bin() string {
	return filepath.Join(VIPBIN, "vips.exe")
}

func HeaderBin() string {
	return filepath.Join(VIPBIN, "vipsheader.exe")
}
