package storage

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/downloader"
	"tools/runtimes/funcs"
	"tools/runtimes/mainsignal"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client   *minio.Client
	bucket   string
	url      string
	ctx      context.Context
	endpoint string
}

func newMinio(cfg Config) (Storage, error) {

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})

	if err != nil {
		return nil, err
	}
	ctx := cfg.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return &minioStorage{
		client:   client,
		bucket:   cfg.Bucket,
		url:      cfg.URL,
		ctx:      ctx,
		endpoint: cfg.Endpoint,
	}, nil
}

func (m *minioStorage) GetObject(src string) (*url.URL, error) {
	return m.client.PresignedGetObject(
		m.ctx,
		m.bucket,
		src,
		time.Hour,
		nil,
	)
}

func (m *minioStorage) Put(r io.Reader, fm *FileMeta) (string, error) {

	if fm == nil {
		fms, err := PrepareFile(r)
		if err != nil {
			return "", err
		}
		fm = fms
	}

	_, err := m.client.PutObject(
		m.ctx,
		m.bucket,
		fm.ObjectKey,
		fm.Reader,
		fm.Size,
		minio.PutObjectOptions{
			ContentType: fm.ContentType,
		},
	)

	if err != nil {
		return "", err
	}

	hash := hex.EncodeToString(fm.H.Sum(nil))
	objectKey := buildObjectKey(hash, fm.Ext)

	if _, err = m.client.CopyObject(m.ctx, minio.CopyDestOptions{
		Bucket: m.bucket,
		Object: objectKey,
	}, minio.CopySrcOptions{
		Bucket: m.bucket,
		Object: fm.ObjectKey,
	}); err != nil {
		return "", err
	}

	_ = m.client.RemoveObject(m.ctx, m.bucket, fm.ObjectKey, minio.RemoveObjectOptions{})

	return objectKey, err
}

func (m *minioStorage) PutStr(str string) (string, error) {

	str = strings.ReplaceAll(str, "\\", "/")
	if _, err := os.Stat(str); err != nil {
		return "", err
	}

	f, err := os.Open(str)
	if err != nil {
		return "", err
	}

	var fm *FileMeta
	if ffm, err := PrepareFileStr(str, f); err == nil {
		fm = ffm
	}

	name, err := m.Put(f, fm)
	if err != nil {
		f.Close()
		return "", err
	}
	f.Close()
	return name, os.Remove(str)
}

func (m *minioStorage) Get(path string) (io.ReadCloser, error) {

	return m.client.GetObject(
		m.ctx,
		m.bucket,
		path,
		minio.GetObjectOptions{},
	)
}

func (m *minioStorage) Delete(path string) error {

	return m.client.RemoveObject(
		m.ctx,
		m.bucket,
		path,
		minio.RemoveObjectOptions{},
	)
}

func (m *minioStorage) Exists(path string) (bool, error) {

	_, err := m.client.StatObject(
		m.ctx,
		m.bucket,
		path,
		minio.StatObjectOptions{},
	)

	if err != nil {
		return false, nil
	}

	return true, nil
}

func (m *minioStorage) Move(src, dst string) error {

	srcOpt := minio.CopySrcOptions{
		Bucket: m.bucket,
		Object: src,
	}

	dstOpt := minio.CopyDestOptions{
		Bucket: m.bucket,
		Object: dst,
	}

	_, err := m.client.CopyObject(context.Background(), dstOpt, srcOpt)
	if err != nil {
		return err
	}

	return m.Delete(src)
}

func (m *minioStorage) Copy(src, dst string) error {

	srcOpt := minio.CopySrcOptions{
		Bucket: m.bucket,
		Object: src,
	}

	dstOpt := minio.CopyDestOptions{
		Bucket: m.bucket,
		Object: dst,
	}

	_, err := m.client.CopyObject(context.Background(), dstOpt, srcOpt)

	return err
}

func (m *minioStorage) URL(object string) string {
	// return m.url + "/" + path
	object = strings.TrimPrefix(object, "/")
	if object == "" {
		return ""
	}
	if strings.HasPrefix(object, "http") {
		return object
	} else if _, err := os.Stat(object); err == nil {
		return object
	}

	u := url.URL{
		Scheme: "http",
		Host:   m.endpoint,
		Path:   path.Join(m.bucket, object),
	}

	return u.String()
}

func (m *minioStorage) Download(ctx context.Context, url string, opt *downloader.DownloadOption) (string, int64, int64, string, error) {
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

	rsp, err := downloader.Download(ctx, opt)
	if err != nil {
		return "", 0, 0, "", err
	}

	fullname := config.FullPath(rsp.FullName)

	mime, err := funcs.FileMimeType(fullname)

	rname, err := m.PutStr(fullname)
	if err != nil {
		return "", 0, 0, "", err
	}

	return rname, rsp.Size, rsp.End.Unix() - rsp.Start.Unix(), mime, nil
}

func (m *minioStorage) Base(src string) string {
	prev := fmt.Sprintf("http://%s/%s", m.endpoint, m.bucket)
	return strings.Trim(strings.ReplaceAll(src, prev, ""), "/")
}
