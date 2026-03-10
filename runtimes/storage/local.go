package storage

import (
	"context"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/downloader"
	"tools/runtimes/funcs"
	"tools/runtimes/mainsignal"
)

type localStorage struct {
	basePath string
	baseURL  string
}

func newLocal(cfg Config) (Storage, error) {

	return &localStorage{
		basePath: cfg.LocalPath,
		baseURL:  cfg.LocalURL,
	}, nil
}

func (l *localStorage) full(p string) string {
	return filepath.Join(l.basePath, p)
}

func (l *localStorage) Put(r io.Reader) (string, error) {

	meta, err := PrepareFile(r)
	if err != nil {
		return "", err
	}

	tmpPath := filepath.Join(l.basePath, meta.ObjectKey)
	if err := os.MkdirAll(filepath.Dir(tmpPath), os.ModePerm); err != nil {
		return "", err
	}

	// 5️⃣ 本地写入临时文件
	f, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, meta.Reader); err != nil {
		f.Close()
		return "", err
	}
	f.Close()

	// 6️⃣ 计算 hash 并生成最终路径
	hash := hex.EncodeToString(meta.H.Sum(nil))
	objectKey := buildObjectKey(hash, meta.Ext)
	finalPath := filepath.Join(l.basePath, objectKey)
	if err := os.MkdirAll(filepath.Dir(finalPath), os.ModePerm); err != nil {
		return "", err
	}

	// 7️⃣ 移动临时文件到最终路径
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return "", err
	}
	return objectKey, nil
}

func (l *localStorage) PutStr(str string) (string, error) {
	if _, err := os.Stat(str); err != nil {
		return "", nil
	}

	f, err := os.Open(str)
	if err != nil {
		return "", err
	}

	name, err := l.Put(f)
	if err != nil {
		f.Close()
		return "", err
	}
	f.Close()
	return name, os.Remove(str)
}

func (l *localStorage) Get(path string) (io.ReadCloser, error) {
	return os.Open(l.full(path))
}

func (l *localStorage) Delete(path string) error {
	return os.Remove(l.full(path))
}

func (l *localStorage) Exists(path string) (bool, error) {

	_, err := os.Stat(l.full(path))

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (l *localStorage) Move(src, dst string) error {

	srcPath := l.full(src)
	dstPath := l.full(dst)

	err := os.MkdirAll(filepath.Dir(dstPath), 0755)
	if err != nil {
		return err
	}

	return os.Rename(srcPath, dstPath)
}

func (l *localStorage) Copy(src, dst string) error {

	r, err := os.Open(l.full(src))
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(l.full(dst))
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return err
}

func (l *localStorage) URL(path string) string {
	return l.baseURL + "/" + path
}

func (l *localStorage) Download(ctx context.Context, url string, opt *downloader.DownloadOption) (string, int64, int64, string, error) {

	dirname := config.FullPath(config.MEDIAROOT, ".tmp")
	if opt == nil {
		opt = &downloader.DownloadOption{
			URL:      url,
			Timeout:  time.Second * 120,
			Dir:      dirname,
			MainWait: &mainsignal.MainWait,
		}
	}
	if opt.Dir == "" {
		opt.Dir = dirname
	}
	opt.URL = url
	name, err := downloader.Download(ctx, opt)
	if err != nil {
		return "", 0, 0, "", err
	}

	fullname := config.FullPath(dirname, name.FullName)

	mime, _ := funcs.FileMimeType(fullname)

	rname, err := l.PutStr(fullname)
	if err != nil {
		return "", 0, 0, "", err
	}
	return rname, name.Size, name.End.Unix() - name.Start.Unix(), mime, nil
}
