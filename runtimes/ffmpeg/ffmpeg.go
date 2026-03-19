package ffmpeg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"
	"tools/runtimes/services"
)

func init() {
	ddir := config.FullPath(config.SYSROOT, "ffmpeg")
	if _, err := os.Stat(filepath.Join(ddir, "ffmpeg.exe")); err != nil {
		go func() {
			if err := download(ddir); err != nil {
				logs.Error(err.Error())
				fmt.Println(err)
			} else {
				config.FFmpeg = ddir
			}
		}()
	} else {
		config.FFmpeg = ddir
	}
}

func download(ddir string) error {
	fmt.Println("下载 ffmpeg---")

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
		fmt.Println("ffmpeg 下载失败")
		return err
	}

	fmt.Println("\n下载完成,开始解压......")
	if err := unzip(config.FullPath(ddir, "ffmpeg.zip"), ddir); err != nil {
		return err
	}

	config.FFmpeg = ddir
	return nil
}

func unzip(fullname, zipto string) error {
	if err := funcs.Unzip(fullname, zipto); err != nil {
		logs.Error("解压 ffmpeg 失败:" + err.Error())
		fmt.Println("解压 ffmpeg 失败")
		return err
	}
	os.Remove(fullname)

	// 移动文件
	if err := moveFile(zipto); err != nil {
		return err
	}

	fmt.Println("解压完成!")
	return nil
}

func moveFile(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var binPath string

	// 找到 bin 目录
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		p := filepath.Join(dir, e.Name(), "bin")
		if _, err := os.Stat(p); err == nil {
			binPath = p
			break
		}
	}

	if binPath == "" {
		return fmt.Errorf("bin directory not found")
	}

	files, err := os.ReadDir(binPath)
	if err != nil {
		return err
	}

	// 复制 bin 文件到 ffmpeg 目录
	for _, f := range files {

		src := filepath.Join(binPath, f.Name())
		dst := filepath.Join(dir, f.Name())

		if err := copyFile(src, dst); err != nil {
			return err
		}
	}

	// 删除原来的所有目录
	for _, e := range entries {
		os.RemoveAll(filepath.Join(dir, e.Name()))
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func serverName() string {
	return runtime.GOOS + "/ffmpeg.zip"
}

func getffmpegPath() string {
	str := "ffmpeg"
	switch runtime.GOOS {
	case "windows":
		str = str + ".exe"
	default:
	}
	return config.FullPath(config.SYSROOT, "ffmpeg", str)
}
func getffprobPath() string {
	str := "ffprobe"
	switch runtime.GOOS {
	case "windows":
		str = str + ".exe"
	default:
	}
	return config.FullPath(config.SYSROOT, "ffmpeg", str)
}

// 执行ffmpeg方法
func RunFfmpeg(wait bool, args ...string) (string, *exec.Cmd, error) {
	return funcs.RunCommand(wait, getffmpegPath(), args...)
}

// 执行ffporbe方法
func RunFfporbe(wait bool, args ...string) (string, *exec.Cmd, error) {
	return funcs.RunCommand(wait, getffprobPath(), args...)
}
