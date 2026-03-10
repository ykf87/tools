package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"mime"
	"net/http"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/downloader"

	"github.com/google/uuid"
)

// type DownloadOption struct {

// 	// HTTP header
// 	Headers map[string]string

// 	// cookie
// 	Cookie map[string]string

// 	// 代理
// 	Proxy string

// 	// 超时
// 	Timeout time.Duration

// 	// 进度
// 	OnProgress Progress
// }

type Storage interface {

	// 上传
	Put(reader io.Reader) (string, error)

	// 上传
	PutStr(str string) (string, error)

	// 下载
	Get(path string) (io.ReadCloser, error)

	// 删除
	Delete(path string) error

	// 是否存在
	Exists(path string) (bool, error)

	// 移动
	Move(src, dst string) error

	// 复制
	Copy(src, dst string) error

	// 获取URL
	URL(path string) string

	Download(ctx context.Context, url string, opt *downloader.DownloadOption) (string, int64, int64, string, error)
}

type FileMeta struct {
	Ext       string
	ObjectKey string
	Reader    io.Reader
	H         hash.Hash
}

type sts map[string]Storage

var Storages = sts{}

func Load(key string) Storage {
	if v, ok := Storages[key]; ok {
		return v
	}
	return Storages[config.DefStorage]
}

func init() {
	if s, err := New(Config{
		Type:      "local",
		LocalPath: config.FullPath(config.MEDIAROOT),
		LocalURL:  config.MediaUrl,
	}); err == nil {
		Storages["local"] = s
	}

	if s, err := New(Config{
		Type:      "minio",
		AccessKey: config.ACCESSKEY,
		SecretKey: config.SECRETKEY,
		Bucket:    config.BUCKET,
		Endpoint:  fmt.Sprintf("127.0.0.1:%d", config.MINIAPIPORT),
		UseSSL:    config.USESSL,
		URL:       fmt.Sprintf("http://127.0.0.1:%d", config.MINIPORT),
	}); err == nil {
		Storages["minio"] = s
	}
}

func PrepareFile(r io.Reader) (*FileMeta, error) {
	// 1️⃣ 读取前512字节检测 MIME
	head := make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	head = head[:n]

	mimeType := http.DetectContentType(head)
	ext := mimeToExt(mimeType)

	// 2️⃣ 构建临时 ObjectKey
	tmpKey := fmt.Sprintf("tmp/%s", uuid.New().String())

	// 3️⃣ 构建完整 reader（前512 + 剩余 r）
	fullReader := io.MultiReader(bytes.NewReader(head), r)

	// 4️⃣ 用 TeeReader 边上传边计算 SHA256 hash
	h := sha256.New()
	tr := io.TeeReader(fullReader, h)

	return &FileMeta{
		Ext:       ext,
		ObjectKey: tmpKey,
		Reader:    tr,
		H:         h,
	}, nil
}

// mimeToExt 使用 mime 包解析扩展名
func mimeToExt(mimeType string) string {
	exts, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(exts) == 0 {
		return "bin"
	}

	ext := strings.TrimPrefix(exts[0], ".")

	// 统一常用扩展名
	switch ext {
	case "jpeg":
		return "jpg"
	case "tiff":
		return "tif"
	}

	return ext
}

// buildObjectKey 根据 hash 生成分片目录对象路径
func buildObjectKey(hash, ext string) string {
	if len(hash) < 4 {
		return fmt.Sprintf("%s.%s", hash, ext)
	}
	return fmt.Sprintf("%s/%s/%s.%s",
		hash[:2], hash[2:4], hash, ext)
}
