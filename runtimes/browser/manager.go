package browser

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type Manager struct {
	mu       sync.Mutex
	browsers map[string]*Browser
	baseDir  string
}

func NewManager(baseDir string) *Manager {
	os.MkdirAll(baseDir, 0755)
	return &Manager{
		browsers: make(map[string]*Browser),
		baseDir:  baseDir,
	}
}

func (m *Manager) New(id string, opt Options) (*Browser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if b, ok := m.browsers[id]; ok && !b.closed.Load() {
		return b, nil
	}

	if opt.ExecPath == "" {
		return nil, errors.New("chrome exec path required")
	}

	userDir := opt.UserDir
	if userDir == "" {
		userDir = m.baseDir
	}
	os.MkdirAll(userDir, 0755)

	allocOpts := make([]chromedp.ExecAllocatorOption, 0, len(chromedp.DefaultExecAllocatorOptions)+8)
	allocOpts = append(allocOpts, chromedp.DefaultExecAllocatorOptions[:]...)
	// fmt.Println(opt.ExecPath, userDir)
	allocOpts = append(allocOpts,
		chromedp.ExecPath(opt.ExecPath),
		chromedp.UserDataDir(userDir),
		chromedp.WindowSize(opt.Width, opt.Height),
		chromedp.Flag("headless", opt.Headless),
		chromedp.Flag("disable-gpu", opt.Headless),
		chromedp.Flag("worker-id", id),
	)

	if opt.Proxy != "" {
		allocOpts = append(allocOpts, chromedp.ProxyServer(opt.Proxy))
	}
	if opt.UserAgent != "" {
		allocOpts = append(allocOpts, chromedp.UserAgent(opt.UserAgent))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...any) {}))
	// ✅ 关键：立刻创建一个存活 target
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		allocCancel()
		return nil, err
	}

	b := &Browser{
		id:       id,
		opts:     opt,
		ctx:      ctx,
		cancel:   cancel,
		alloc:    allocCancel,
		onClosed: make(chan struct{}),
	}
	b.onURLChange.Store((func(string))(nil))
	b.onConsole.Store((func([]*runtime.RemoteObject))(nil))
	b.watchClose()
	go b.startEventLoop()

	go func() {
		<-ctx.Done()
		m.remove(id)
	}()

	m.browsers[id] = b
	return b, nil
}

func (m *Manager) remove(id string) {
	m.mu.Lock()
	b, ok := m.browsers[id]
	if ok {
		delete(m.browsers, id)
	}
	m.mu.Unlock()

	if ok {
		b.Close()
	}
}
