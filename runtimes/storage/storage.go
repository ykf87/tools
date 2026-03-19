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
	"net/url"
	"os"
	"path/filepath"
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
	Put(reader io.Reader, fm *FileMeta) (string, error)

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

	Base(src string) string

	GetObject(src string) (*url.URL, error)

	Download(ctx context.Context, url string, opt *downloader.DownloadOption) (string, int64, int64, string, error)
}

type FileMeta struct {
	Ext         string
	ObjectKey   string
	Reader      io.Reader
	H           hash.Hash
	ContentType string
	Size        int64
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

func detectMedia(head []byte) (ext string, contentType string) {
	// MP4 / MOV（ftyp）
	if len(head) > 12 && string(head[4:8]) == "ftyp" {
		return ".mp4", "video/mp4"
	}

	// MP3（ID3 或帧头）
	if bytes.HasPrefix(head, []byte("ID3")) ||
		(len(head) > 2 && head[0] == 0xFF && (head[1]&0xE0) == 0xE0) {
		return ".mp3", "audio/mpeg"
	}

	// WAV
	if bytes.HasPrefix(head, []byte("RIFF")) &&
		bytes.Contains(head, []byte("WAVE")) {
		return ".wav", "audio/wav"
	}

	// AAC（ADTS）
	if len(head) > 2 && head[0] == 0xFF && (head[1]&0xF6) == 0xF0 {
		return ".aac", "audio/aac"
	}

	// fallback（最后兜底）
	mimeType := http.DetectContentType(head)

	switch {
	case strings.HasPrefix(mimeType, "video/"):
		return ".mp4", mimeType
	case strings.HasPrefix(mimeType, "audio/"):
		return ".mp3", mimeType
	}

	return "", "application/octet-stream"
}

func resolveContentType(ext, mime string) string {
	ext = strings.ToLower(ext)

	// ===== 1️⃣ 优先按扩展名判断（最可靠） =====
	switch ext {

	// ===== 视频 =====
	case ".mp4":
		return "video/mp4"
	case ".m4v":
		return "video/x-m4v"
	case ".mkv":
		return "video/x-matroska"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".flv":
		return "video/x-flv"
	case ".webm":
		return "video/webm"
	case ".ts":
		return "video/mp2t"
	case ".3gp":
		return "video/3gpp"

	// ===== 音频 =====
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".aac":
		return "audio/aac"
	case ".m4a":
		return "audio/mp4"
	case ".ogg":
		return "audio/ogg"
	case ".opus":
		return "audio/opus"
	case ".wma":
		return "audio/x-ms-wma"

	// ===== 图片 =====
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".svg":
		return "image/svg+xml"
	case ".tiff", ".tif":
		return "image/tiff"
	case ".ico":
		return "image/x-icon"

	// ===== 文档 =====
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"

	// ===== Office =====
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"

	// ===== 压缩包 =====
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/vnd.rar"
	case ".7z":
		return "application/x-7z-compressed"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	}

	// ===== 2️⃣ fallback: 使用检测出来的 MIME =====
	if mime != "" && mime != "application/octet-stream" {
		// 只接受合理类型
		if strings.HasPrefix(mime, "video/") ||
			strings.HasPrefix(mime, "audio/") ||
			strings.HasPrefix(mime, "image/") ||
			strings.HasPrefix(mime, "text/") ||
			strings.Contains(mime, "json") ||
			strings.Contains(mime, "xml") {
			return mime
		}
	}

	// ===== 3️⃣ 最终 fallback =====
	return "application/octet-stream"
}

func PrepareFile(r io.Reader) (*FileMeta, error) {
	// 1️⃣ 读取前512字节检测 MIME
	head := make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	head = head[:n]

	// mimeType := http.DetectContentType(head)
	// ext := mimeToExt(mimeType)

	// 2️⃣ 构建临时 ObjectKey
	tmpKey := fmt.Sprintf("tmp/%s", uuid.New().String())

	// 3️⃣ 构建完整 reader（前512 + 剩余 r）
	fullReader := io.MultiReader(bytes.NewReader(head), r)

	// 4️⃣ 用 TeeReader 边上传边计算 SHA256 hash
	h := sha256.New()
	tr := io.TeeReader(fullReader, h)

	ext, contentType := detectMedia(head)

	return &FileMeta{
		Ext:         ext,
		ObjectKey:   tmpKey,
		Reader:      tr,
		H:           h,
		Size:        -1,
		ContentType: contentType,
	}, nil
}

func PrepareFileStr(filename string, file *os.File) (*FileMeta, error) {
	// 1️⃣ 打开文件
	// file, err := os.Open(filename)
	// if err != nil {
	// 	return nil, err
	// }

	// 2️⃣ 获取文件信息（拿 size）
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	size := stat.Size()

	// 3️⃣ 读取头部用于检测
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		file.Close()
		return nil, err
	}
	head = head[:n]

	// 4️⃣ 重置文件指针（关键！）
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, err
	}

	// 5️⃣ 类型识别（优先 ext + fallback 文件头）
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := http.DetectContentType(head)

	contentType := resolveContentType(ext, mimeType)

	// 6️⃣ hash
	h := sha256.New()
	tr := io.TeeReader(file, h)

	// 7️⃣ 临时 key（如果你还需要）
	tmpKey := fmt.Sprintf("tmp/%s", uuid.New().String())
	if size < 1 {
		size = -1
	}

	return &FileMeta{
		Ext:         ext,
		ObjectKey:   tmpKey,
		Reader:      tr,
		H:           h,
		Size:        size,
		ContentType: contentType,
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
	ext = strings.Trim(ext, ".")
	if len(hash) < 4 {
		return fmt.Sprintf("%s.%s", hash, ext)
	}
	return fmt.Sprintf("%s/%s/%s.%s",
		hash[:2], hash[2:4], hash, ext)
}
