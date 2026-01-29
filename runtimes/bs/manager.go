package bs

import (
	"errors"
	"fmt"
	"sync"

	rt "github.com/chromedp/cdproto/runtime"
)

// 控制器
type Manager struct {
	mu       sync.Mutex
	browsers map[int64]*Browser
	baseDir  string
}

// 临时的浏览器id
var TempIndex int64
var BsManager *Manager // 由于浏览器的特殊性,在一次打开后本质上是不允许在这次打开再操作的,因此统一管理

func init() {
	BsManager = newManager("")
}

// 新建一个浏览器组
func newManager(baseDir string) *Manager {
	if baseDir == "" {
		baseDir = BASEPATH
	}
	return &Manager{
		browsers: make(map[int64]*Browser),
		baseDir:  baseDir,
	}
}

// 仅对控制器执行增减操作,并不启动浏览器
func (m *Manager) New(id int64, opt Options, wait bool) (*Browser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id < 1 {
		TempIndex--
		id = TempIndex
		opt.Temp = true
	}

	if b, ok := m.browsers[id]; ok {
		return b, nil
	}
	if b, ok := OpendBrowser.Load(id); ok {
		if bbs, ok := b.(*Browser); ok {
			if wait == true {
				if bbs.Locker != nil {
					<-bbs.Locker
				}
				return bbs, nil
			} else {
				return nil, fmt.Errorf("浏览器已经打开")
			}
		}
	}

	if opt.ExecPath == "" {
		execPath, err := getBrowserBinName()
		if err != nil {
			return nil, errors.New("chrome exec path required")
		}
		opt.ExecPath = execPath
	}

	if opt.UserDir == "" {
		userDir, err := GetBrowserConfigDir(id)
		if err != nil {
			return nil, err
		}
		opt.UserDir = userDir
	}
	if opt.Width == 0 {
		opt.Width = 800
	}
	if opt.Height == 0 {
		opt.Height = 600
	}

	b := &Browser{
		ID:   id,
		Opts: opt,
	}

	if _, err := MakeBrowserConfig(b.ID, b.Opts.Language, b.Opts.Timezone, b.Opts.Proxy); err != nil {
		return nil, err
	}

	b.onURLChange.Store((func(string))(nil))
	b.onConsole.Store((func([]*rt.RemoteObject))(nil))

	m.browsers[id] = b
	OpendBrowser.Store(id, b)
	return b, nil
}

func (m *Manager) Close(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bs, ok := m.browsers[id]
	if !ok {
		return nil
	}
	if bs.survival.Load() {
		bs.Close()
	}
	return nil
}

// 删除当前manager的一个浏览器
func (m *Manager) Remove(id int64) {
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

// 是否存活
func (m *Manager) IsArride(id int64) bool {
	if bs, ok := m.browsers[id]; ok {
		return bs.survival.Load()
	}
	return false
}
