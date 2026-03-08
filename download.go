package main

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

type Callback func(total, downloaded, speed, workers int64)

type DownloadOption struct {
	URL      string
	Dir      string
	FileName string

	Threads int

	Proxy   string
	Headers map[string]string
	Cookies map[string]string

	Callback Callback
}

type chunk struct {
	start int64
	end   int64
}

type writeTask struct {
	offset int64
	data   []byte
}

func newClient(proxy string) *http.Client {

	tr := &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
	}

	if proxy != "" {
		p, _ := url.Parse(proxy)
		tr.Proxy = http.ProxyURL(p)
	}

	jar, _ := cookiejar.New(nil)

	return &http.Client{
		Transport: tr,
		Jar:       jar,
	}
}

func resolveFileName(resp *http.Response, u string) string {

	cd := resp.Header.Get("Content-Disposition")

	if cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if v, ok := params["filename"]; ok {
				return v
			}
		}
	}

	pu, _ := url.Parse(u)
	name := path.Base(pu.Path)

	if name != "" && name != "/" {
		return name
	}

	return fmt.Sprintf("%d.bin", time.Now().Unix())
}

func applyHeaders(req *http.Request, h map[string]string) {

	for k, v := range h {
		req.Header.Set(k, v)
	}
}

func writer(
	ctx context.Context,
	file *os.File,
	ch chan writeTask,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	for {

		select {

		case <-ctx.Done():
			return

		case t, ok := <-ch:

			if !ok {
				return
			}

			file.WriteAt(t.data, t.offset)

		}

	}
}

func worker(
	ctx context.Context,
	client *http.Client,
	url string,
	headers map[string]string,
	chunkQ chan chunk,
	writeCh chan writeTask,
	progress *int64,
	active *int64,
) {

	atomic.AddInt64(active, 1)
	defer atomic.AddInt64(active, -1)

	buf := make([]byte, 128*1024)

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

			req.Header.Set(
				"Range",
				fmt.Sprintf("bytes=%d-%d", c.start, c.end),
			)

			resp, err := client.Do(req)

			if err != nil {
				continue
			}

			offset := c.start

			for {

				n, err := resp.Body.Read(buf)

				if n > 0 {

					data := make([]byte, n)
					copy(data, buf[:n])

					writeCh <- writeTask{
						offset: offset,
						data:   data,
					}

					offset += int64(n)

					atomic.AddInt64(progress, int64(n))
				}

				if err != nil {
					resp.Body.Close()
					break
				}

			}

		}

	}

}

func speedMonitor(
	ctx context.Context,
	total int64,
	progress *int64,
	active *int64,
	cb Callback,
) {

	ticker := time.NewTicker(time.Second)

	var last int64

	for {

		select {

		case <-ctx.Done():
			return

		case <-ticker.C:

			cur := atomic.LoadInt64(progress)
			spd := cur - last
			last = cur

			workers := atomic.LoadInt64(active)

			if cb != nil {
				cb(total, cur, spd, workers)
			}

		}

	}

}

func DownloadFile(ctx context.Context, opt DownloadOption) error {

	if opt.Threads <= 0 {
		opt.Threads = 4
	}

	client := newClient(opt.Proxy)

	req, _ := http.NewRequestWithContext(ctx, "GET", opt.URL, nil)

	applyHeaders(req, opt.Headers)

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	size := resp.ContentLength

	if size <= 0 {
		return fmt.Errorf("invalid file size")
	}

	name := opt.FileName
	if name == "" {
		name = resolveFileName(resp, opt.URL)
	}

	resp.Body.Close()

	filePath := path.Join(opt.Dir, name)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	file.Truncate(size)

	chunkSize := size / int64(opt.Threads)

	queue := make(chan chunk, opt.Threads)

	for i := 0; i < opt.Threads; i++ {

		start := int64(i) * chunkSize
		end := start + chunkSize - 1

		if i == opt.Threads-1 {
			end = size - 1
		}

		queue <- chunk{start, end}

	}

	close(queue)

	writeCh := make(chan writeTask, 1024)

	var writerWG sync.WaitGroup
	writerWG.Add(1)

	go writer(ctx, file, writeCh, &writerWG)

	var progress int64
	var active int64

	for i := 0; i < opt.Threads; i++ {

		go worker(
			ctx,
			client,
			opt.URL,
			opt.Headers,
			queue,
			writeCh,
			&progress,
			&active,
		)

	}

	go speedMonitor(ctx, size, &progress, &active, opt.Callback)

	for {

		select {

		case <-ctx.Done():
			return ctx.Err()

		default:

			if atomic.LoadInt64(&progress) >= size {

				close(writeCh)

				writerWG.Wait()

				return nil
			}

			time.Sleep(time.Second)

		}

	}

}
