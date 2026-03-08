package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
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

func (l *localStorage) Put(path string, r io.Reader) error {

	full := l.full(path)

	err := os.MkdirAll(filepath.Dir(full), 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
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

func (l *localStorage) Download(ctx context.Context, url string, opt DownloadOption) (string, error) {

	name := opt.FileName

	if name == "" {
		name = uuid.New().String()
	}

	tmp := filepath.Join(os.TempDir(), name+".downloading")

	f, err := os.Create(tmp)

	if err != nil {
		return "", err
	}

	resp, err := StreamDownload(url, opt, f)

	f.Close()

	if err != nil {
		return "", err
	}

	ext := ExtFromURL(url)

	if ext == "" {
		ext = ExtFromHeader(resp)
	}

	if ext == "" {
		ext = ".bin"
	}

	final := name + ext

	path := filepath.Join(opt.Dir, final)

	full := l.full(path)

	os.MkdirAll(filepath.Dir(full), 0755)

	err = os.Rename(tmp, full)

	return path, err
}
