package downloader

import (
	"context"
	"fmt"
	"math/rand"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"tools/runtimes/mainsignal"
)

var UAS = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:125.0) Gecko/20100101 Firefox/125.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:126.0) Gecko/20100101 Firefox/126.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:148.0) Gecko/20100101 Firefox/148.0",

	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; ARM Mac OS X 14_3_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; ARM Mac OS X 14_2_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",

	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_6_3) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_3_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",

	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.6; rv:124.0) Gecko/20100101 Firefox/124.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.6; rv:125.0) Gecko/20100101 Firefox/125.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.0; rv:126.0) Gecko/20100101 Firefox/126.0",
	"Mozilla/5.0 (Macintosh; ARM Mac OS X 14.3; rv:125.0) Gecko/20100101 Firefox/125.0",
}

type Callback func(total, downloaded, speed, workers int64)

type DownloadOption struct {
	URL      string
	Dir      string
	FileName string
	Threads  int

	Proxy   string
	Headers map[string]string
	Cookies map[string]string

	Callback Callback
	MainWait *sync.WaitGroup

	Timeout time.Duration
}

type chunk struct {
	start int64
	end   int64
}

type writeTask struct {
	offset int64
	data   []byte
}

type DownLoadFileInfo struct {
	Start    time.Time
	End      time.Time
	Size     int64
	Dir      string
	Name     string
	FullName string
	Ext      string
}

var bufPool = sync.Pool{
	New: func() any { return make([]byte, 128*1024) },
}

func newClient(proxy string, timeout time.Duration) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
	}
	if proxy != "" {
		p, _ := url.Parse(proxy)
		tr.Proxy = http.ProxyURL(p)
	}
	jar, _ := cookiejar.New(nil)

	if timeout < 1 {
		timeout = time.Second * 120
	}

	return &http.Client{
		Transport: tr,
		Jar:       jar,
		Timeout:   timeout,
	}
}

func resolveFileName(resp *http.Response, u string) (string, string) {
	var name, ext string

	// 1. 尝试从 Content-Disposition 中获取文件名
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if v, ok := params["filename"]; ok && v != "" {
				name = v
			}
		}
	}

	// 2. 如果没有，从 URL 路径获取
	if name == "" {
		pu, _ := url.Parse(u)
		name = path.Base(pu.Path)
	}

	// 3. 如果仍然为空或为斜杠，生成默认名字
	if name == "" || name == "/" {
		name = fmt.Sprintf("%d", time.Now().Unix())
	}

	// 4. 尝试根据 Content-Type 推断扩展名
	if !strings.Contains(name, ".") {
		ct := resp.Header.Get("Content-Type")
		if ct != "" {
			if exts, err := mime.ExtensionsByType(ct); err == nil && len(exts) > 0 {
				ext = exts[0]
			}
		}
	}

	// 5. 最终拼接扩展名（如果已经有则保持原名）
	if ext != "" && filepath.Ext(name) == "" {
		name += ext
	}

	return name, ext
}

func applyHeaders(req *http.Request, h map[string]string) {
	if h == nil {
		h = make(map[string]string)
	}
	if _, ok := h["User-Agent"]; !ok {
		h["User-Agent"] = UAS[rand.Intn(len(UAS))]
	}
	for k, v := range h {
		req.Header.Set(k, v)
	}
}

// 写文件 goroutine
func writer(ctx context.Context, file *os.File, ch <-chan writeTask, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			// ctx 取消，直接返回
			return
		case t, ok := <-ch:
			if !ok {
				// channel 关闭，退出
				return
			}
			if _, err := file.WriteAt(t.data, t.offset); err != nil {
				fmt.Println("write error:", err)
				return
			}
		}
	}
}

// worker 下载 goroutine
func worker(ctx context.Context, client *http.Client, url string, headers map[string]string, chunkQ <-chan chunk, writeCh chan<- writeTask, progress *int64, active *int64) {
	atomic.AddInt64(active, 1)
	defer atomic.AddInt64(active, -1)

	// buf := make([]byte, 128*1024)
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)
	for {
		select {
		case <-ctx.Done():
			return
		case c, ok := <-chunkQ:
			if !ok {
				return
			}
			req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
			applyHeaders(req, headers)
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", c.start, c.end))

			// resp, err := client.Do(req)
			// if err != nil || resp.StatusCode >= 400 {
			// 	fmt.Println("request error:", err, resp.Status)
			// 	continue
			// }
			var resp *http.Response
			for retry := range [3]int{} {
				rp, err := client.Do(req)
				if err == nil && rp.StatusCode < 400 {
					resp = rp
					break
				}
				time.Sleep(time.Second)
				if retry == 2 {
					return
					// 最终失败
					// fmt.Println("failed chunk:", c)
				}
			}
			if resp == nil {
				fmt.Println("failed chunk:", c)
				continue
			}
			defer resp.Body.Close()

			offset := c.start
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					data := make([]byte, n)
					copy(data, buf[:n])
					select {
					case writeCh <- writeTask{offset: offset, data: data}:
						offset += int64(n)
						atomic.AddInt64(progress, int64(n))
					case <-ctx.Done():
						// resp.Body.Close()
						return
					}
				}
				if err != nil {
					// resp.Body.Close()
					break
				}
			}
		}
	}
}

// 速度监控 goroutine
func speedMonitor(ctx context.Context, total int64, progress *int64, active *int64, cb Callback) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var last int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if total <= 0 {
				continue // 或者 total = 1 防止除零
			}
			cur := atomic.LoadInt64(progress)
			spd := cur - last
			last = cur
			if cb != nil {
				go cb(total, cur, spd, atomic.LoadInt64(active))
			}
		}
	}
}

// 辅助函数，获取响应状态（nil安全）
func respStatus(resp *http.Response) string {
	if resp == nil {
		return "nil"
	}
	return fmt.Sprintf("%d %s", resp.StatusCode, resp.Status)
}

func getFileSize(ctx context.Context, client *http.Client, url string, headers map[string]string) (int64, error) {
	// --- 1. 尝试 HEAD ---
	req, _ := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	applyHeaders(req, headers)
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode < 400 {
		defer resp.Body.Close()
		if resp.ContentLength > 0 {
			return resp.ContentLength, nil
		}
	}

	// --- 2. 降级 GET bytes=0-0 ---
	req2, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	applyHeaders(req2, headers)
	req2.Header.Set("Range", "bytes=0-0")
	resp2, err := client.Do(req2)
	if err != nil {
		return 0, fmt.Errorf("GET range request failed: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode >= 400 {
		return 0, fmt.Errorf("GET range request failed with status %d", resp2.StatusCode)
	}

	// Content-Range 格式: bytes 0-0/12345
	cr := resp2.Header.Get("Content-Range")
	if cr != "" {
		parts := strings.Split(cr, "/")
		if len(parts) == 2 {
			if total, err := strconv.ParseInt(parts[1], 10, 64); err == nil && total > 0 {
				return total, nil
			}
		}
	}

	// --- 3. 都失败 → 返回错误，带服务端响应信息 ---
	return 0, fmt.Errorf("cannot determine file size, HEAD status: %v, GET status: %v", respStatus(resp), respStatus(resp2))
}

// func resolveFileNameFromResp(resp *http.Response, u string) (string, string) {
// 	name, ext := "", ""

// 	// 1. 尝试 Content-Disposition
// 	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
// 		if _, params, err := mime.ParseMediaType(cd); err == nil {
// 			if v, ok := params["filename"]; ok && v != "" {
// 				name = v
// 			}
// 		}
// 	}

// 	// 2. URL 路径
// 	if name == "" {
// 		pu, _ := url.Parse(u)
// 		name = path.Base(pu.Path)
// 	}

// 	// 3. 默认名字
// 	if name == "" || name == "/" {
// 		name = fmt.Sprintf("%d", time.Now().Unix())
// 	}

// 	// 4. 根据 Content-Type 推断扩展名
// 	if !strings.Contains(name, ".") {
// 		ct := resp.Header.Get("Content-Type")
// 		if ct != "" {
// 			if exts, err := mime.ExtensionsByType(ct); err == nil && len(exts) > 0 {
// 				ext = exts[0]
// 			}
// 		}
// 	}

// 	if ext != "" && filepath.Ext(name) == "" {
// 		name += ext
// 	}

// 	return name, ext
// }

// 外部调用接口
// 使用方式:
//
//	name, err := downloader.Download(mainsignal.MainCtx, &downloader.DownloadOption{
//		URL:      "download url",
//		Dir:      "./",
//		FileName: "",
//		Threads:  8,
//		// Headers: map[string]string{
//		// 	"User-Agent": "Mozilla/5.0",
//		// },
//		Callback: func(total, cur, speed, workers int64) {
//			fmt.Printf(
//				"\r%.2f%% %s/s workers:%d %s",
//				float64(cur)/float64(total)*100,
//				funcs.FormatFileSize(speed, "1", ""),
//				workers,
//				funcs.FormatFileSize(total, "1", ""),
//			)
//		},
//	})
func Download(ctx context.Context, opt *DownloadOption) (*DownLoadFileInfo, error) {
	if opt.MainWait == nil {
		opt.MainWait = &mainsignal.MainWait
	}
	opt.MainWait.Add(1)
	defer opt.MainWait.Done()
	return df(ctx, opt)
}

func safeFileName(name string) string {
	invalid := regexp.MustCompile(`[\\/:*?"<>|]`)

	name = invalid.ReplaceAllString(name, "_")

	return name
}

func df(ctx context.Context, opt *DownloadOption) (*DownLoadFileInfo, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if opt.Threads <= 0 {
		opt.Threads = 4
	}

	client := newClient(opt.Proxy, opt.Timeout)

	if len(opt.Cookies) > 0 {
		u, _ := url.Parse(opt.URL)
		var cookies []*http.Cookie
		for k, v := range opt.Cookies {
			cookies = append(cookies, &http.Cookie{Name: k, Value: v})
		}
		client.Jar.SetCookies(u, cookies)
	}

	size, name, err := func() (int64, string, error) {
		size, err := getFileSize(ctx, client, opt.URL, opt.Headers)
		if err != nil || size <= 0 {
			return 0, "", fmt.Errorf("invalid file size: %v", err)
		}

		// 获取文件名
		req, _ := http.NewRequestWithContext(ctx, "GET", opt.URL, nil)
		applyHeaders(req, opt.Headers)
		resp, err := client.Do(req)
		if err != nil {
			return 0, "", err
		}
		defer resp.Body.Close()

		fileName := opt.FileName
		if fileName == "" {
			fileName, _ = resolveFileName(resp, opt.URL)
		} else if !strings.Contains(fileName, ".") {
			_, ext := resolveFileName(resp, opt.URL)
			fileName = fmt.Sprintf("%s%s", fileName, ext)
		}
		return size, safeFileName(fileName), nil
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	filePath := path.Join(opt.Dir, name)
	partPath := filePath + ".part"

	if _, err := os.Stat(partPath); err == nil {
		return nil, fmt.Errorf("file is downloading: %s", opt.URL)
	}

	if _, err := os.Stat(opt.Dir); err != nil {
		if err := os.MkdirAll(opt.Dir, os.ModePerm); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(partPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	complated := false
	defer func() {
		if !complated {
			file.Close()
			os.Remove(partPath)
		}
	}()
	if err := file.Truncate(size); err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	startTime := time.Now()

	chunkQ := make(chan chunk, opt.Threads)
	go func() {
		chunkSize := size / int64(opt.Threads)
		for i := 0; i < opt.Threads; i++ {
			start := int64(i) * chunkSize
			end := start + chunkSize - 1
			if i == opt.Threads-1 {
				end = size - 1
			}
			chunkQ <- chunk{start: start, end: end}
		}
		close(chunkQ)
	}()

	writeCh := make(chan writeTask, 1024)
	var writerWG sync.WaitGroup
	writerWG.Add(1)
	go writer(ctx, file, writeCh, &writerWG)

	var progress int64
	var active int64

	// 启动 worker
	var workerWG sync.WaitGroup
	workerWG.Add(opt.Threads)
	for i := 0; i < opt.Threads; i++ {
		go func() {
			defer workerWG.Done()
			worker(ctx, client, opt.URL, opt.Headers, chunkQ, writeCh, &progress, &active)
		}()
	}

	// 关闭 writeCh 的 goroutine
	go func() {
		workerWG.Wait()
		close(writeCh)
	}()

	// 启动速度监控
	go speedMonitor(ctx, size, &progress, &active, opt.Callback)

	// 等待下载完成或 ctx 取消
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if atomic.LoadInt64(&progress) >= size {
				file.Close()
				writerWG.Wait()
				complated = true

				if err := os.Rename(partPath, filePath); err != nil {
					return nil, err
				}

				return &DownLoadFileInfo{
					Size:     size,
					FullName: filePath,
					Name:     filepath.Base(filePath),
					Ext:      filepath.Ext(filePath),
					Dir:      opt.Dir,
					End:      time.Now(),
					Start:    startTime,
				}, nil
			}
		}
	}
}
