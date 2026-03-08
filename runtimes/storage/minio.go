package storage

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client *minio.Client
	bucket string
	url    string
	ctx    context.Context
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
		client: client,
		bucket: cfg.Bucket,
		url:    cfg.URL,
		ctx:    ctx,
	}, nil
}

func (m *minioStorage) Put(path string, r io.Reader) error {

	_, err := m.client.PutObject(
		m.ctx,
		m.bucket,
		path,
		r,
		-1,
		minio.PutObjectOptions{},
	)

	return err
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

func (m *minioStorage) URL(path string) string {
	return m.url + "/" + path
}

func (m *minioStorage) Download(ctx context.Context, url string, opt DownloadOption) (string, error) {

	name := opt.FileName

	if name == "" {
		name = uuid.New().String()
	}

	pr, pw := io.Pipe()

	var resp *http.Response
	var err error

	go func() {

		resp, err = StreamDownload(url, opt, pw)

		pw.CloseWithError(err)

	}()

	ext := ExtFromURL(url)

	if ext == "" && resp != nil {
		ext = ExtFromHeader(resp)
	}

	if ext == "" {
		ext = ".bin"
	}

	object := opt.Dir + "/s" + name + ext

	_, err = m.client.PutObject(
		ctx,
		m.bucket,
		object,
		pr,
		-1,
		minio.PutObjectOptions{},
	)

	return object, err
}
