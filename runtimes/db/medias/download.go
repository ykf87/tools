package medias

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/downloader"
	"tools/runtimes/file"
	"tools/runtimes/funcs"
)

// 保留能存入数据库的媒体名称
func MediaBaseName(name string) string {
	name, _ = strings.CutPrefix(name, config.RuningRoot)
	name, _ = strings.CutPrefix(name, config.MEDIAROOT)
	return name
}

// 获取完整的路径名称
func MediaFullName(name string) string {
	name = MediaBaseName(name)
	return filepath.Join(config.RuningRoot, config.MEDIAROOT, name)
}

// 获取媒体的url地址
func MediaUrlName(name string) string {
	name = MediaBaseName(name)
	return fmt.Sprintf("%s%s", config.MediaUrl, name)

}

func DownLoadVideo(urlstr, dir, saveName, proxy string, execing func(percent float64, downloaded, total int64)) (*Media, error) {
	d := downloader.NewDownloader(proxy, execing, nil)

	if saveName == "" {
		saveName = funcs.Md5String(urlstr)
	}
	if strings.Contains(saveName, ".") == false {
		ext, err := d.GetUrlFileExt(urlstr)
		if err != nil {
			return nil, err
		}
		saveName = fmt.Sprintf("%s.%s", saveName, ext)
	}

	if strings.HasPrefix(dir, config.MEDIAROOT) {
		dir, _ = strings.CutPrefix(dir, config.MEDIAROOT)
	}

	fullName := config.FullPath(config.MEDIAROOT, dir, saveName)
	fullDir := filepath.Dir(fullName)
	if _, err := os.Stat(fullDir); err != nil {
		if err := os.MkdirAll(fullDir, os.ModePerm); err != nil {
			return nil, err
		}
	}

	if err := d.Download(urlstr, fullName); err != nil {
		return nil, err
	}

	sn := filepath.Join(dir, saveName)
	md := new(Media)
	fullFn := MediaFullName(sn)
	fl, err := file.NewFileInfo(fullFn)
	if err != nil {
		return nil, err
	}
	md.Mime = fl.GetMime()
	md.Size = fl.Size()
	md.Filetime = fl.Time().Unix()
	md.Md5 = fl.Md5()
	md.Addtime = time.Now()
	md.Name = saveName
	md.Path = dir
	md.Url = urlstr
	md.UrlMd5 = funcs.Md5String(urlstr)

	return md, nil
}
