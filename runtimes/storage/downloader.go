package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Progress func(percent float64, current int64, total int64)

type progressWriter struct {
	writer io.Writer

	total int64

	current int64

	cb Progress
}

func (p *progressWriter) Write(b []byte) (int, error) {

	n, err := p.writer.Write(b)

	p.current += int64(n)

	if p.cb != nil && p.total > 0 {

		percent := float64(p.current) / float64(p.total) * 100

		p.cb(percent, p.current, p.total)
	}

	return n, err
}

func NewHTTPClient(opt DownloadOption) *http.Client {

	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	if opt.Proxy != "" {

		p, err := url.Parse(opt.Proxy)

		if err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	return &http.Client{
		Transport: tr,
		Timeout:   opt.Timeout,
	}
}

func StreamDownload(urlStr string, opt DownloadOption, writer io.Writer) (*http.Response, error) {

	client := NewHTTPClient(opt)

	req, err := http.NewRequest("GET", urlStr, nil)

	if err != nil {
		return nil, err
	}

	if opt.Headers != nil {
		req.Header = opt.Headers
	}

	if opt.Cookie != "" {
		req.Header.Set("Cookie", opt.Cookie)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http error %d", resp.StatusCode)
	}

	pw := &progressWriter{
		writer: writer,
		total:  resp.ContentLength,
		cb:     opt.OnProgress,
	}

	_, err = io.Copy(pw, resp.Body)

	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func RandomName() string {

	b := make([]byte, 16)

	rand.Read(b)

	return hex.EncodeToString(b)
}

func ExtFromURL(u string) string {

	ext := path.Ext(u)

	if len(ext) > 0 && len(ext) < 10 {
		return ext
	}

	return ""
}

func ExtFromHeader(resp *http.Response) string {

	contentType := resp.Header.Get("Content-Type")

	contentType = strings.ToLower(contentType)

	switch {

	case strings.Contains(contentType, "png"):
		return ".png"

	case strings.Contains(contentType, "jpeg"):
		return ".jpg"

	case strings.Contains(contentType, "jpg"):
		return ".jpg"

	case strings.Contains(contentType, "gif"):
		return ".gif"

	case strings.Contains(contentType, "webp"):
		return ".webp"

	case strings.Contains(contentType, "bmp"):
		return ".bmp"

	case strings.Contains(contentType, "svg"):
		return ".svg"

	case strings.Contains(contentType, "mp4"):
		return ".mp4"

	case strings.Contains(contentType, "webm"):
		return ".webm"

	case strings.Contains(contentType, "mkv"):
		return ".mkv"

	case strings.Contains(contentType, "mov"):
		return ".mov"

	case strings.Contains(contentType, "avi"):
		return ".avi"

	case strings.Contains(contentType, "mp3"):
		return ".mp3"

	case strings.Contains(contentType, "wav"):
		return ".wav"

	case strings.Contains(contentType, "aac"):
		return ".aac"

	case strings.Contains(contentType, "flac"):
		return ".flac"

	case strings.Contains(contentType, "pdf"):
		return ".pdf"

	case strings.Contains(contentType, "zip"):
		return ".zip"

	case strings.Contains(contentType, "rar"):
		return ".rar"

	case strings.Contains(contentType, "7z"):
		return ".7z"

	case strings.Contains(contentType, "json"):
		return ".json"

	case strings.Contains(contentType, "text"):
		return ".txt"

	case strings.Contains(contentType, "html"):
		return ".html"

	}

	return ""
}
