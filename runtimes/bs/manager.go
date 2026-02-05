package bs

import (
	"errors"
	"fmt"
	"sync"

	rt "github.com/chromedp/cdproto/runtime"
	"github.com/tidwall/gjson"
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
func (m *Manager) New(id int64, opt *Options, wait bool) (*Browser, error) {
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
		opt.Width = 1024
	}
	if opt.Height == 0 {
		opt.Height = 960
	}

	b := &Browser{
		ID:   id,
		Opts: opt,
	}

	if _, err := MakeBrowserConfig(b.ID, b.Opts.Language, b.Opts.Timezone, b.Opts.Proxy); err != nil {
		return nil, err
	}

	// b.onURLChange.Store((func(string))(nil))
	// b.onConsole.Store((func([]*rt.RemoteObject))(nil))
	//
	b.onURLChange.Store(func(url string) {
		if b.Opts.JsStr != "" {
			b.RunJs(b.Opts.JsStr)
		}
	})
	b.onConsole.Store(func(args []*rt.RemoteObject) {
		for _, arg := range args {
			if arg.Value != nil {
				gs := gjson.Parse(gjson.Parse(arg.Value.String()).String())
				if gs.Get("version").String() == "" {
					continue
				}
				switch gs.Get("type").String() {
				case "success":
					if b.Opts.Msg != nil {
						select {
						case b.Opts.Msg <- gs.Get("data").String():
						case <-b.ctx.Done():
						}
					}
					// fmt.Println("执行成功了啊")
					// b.Close()
				case "fail":
					if b.Opts.Msg != nil {
						select {
						case b.Opts.Msg <- gs.Get("data").String():
						case <-b.ctx.Done():
						}
					}
					// fmt.Println("执行失败了啊")
					// b.Close()
				case "notify":
					if b.Opts.Msg != nil {
						select {
						case b.Opts.Msg <- gs.Get("data").String():
						case <-b.ctx.Done():
						}
					}
				case "upload": // 调用系统的上传功能
					var fls []string
					for _, v := range gs.Get("data.files").Array() {
						fls = append(fls, v.String())
					}
					b.Upload(fls, gs.Get("data.node").String(), gs.Get("data.upnode").String())
					if b.Opts.Msg != nil {
						select {
						case b.Opts.Msg <- "上传文件":
						case <-b.ctx.Done():
						}
					}
				case "input": // 输入
					b.InputTxt(gs.Get("data.text").String(), gs.Get("data.node").String())
					if b.Opts.Msg != nil {
						select {
						case b.Opts.Msg <- "输入数据":
						case <-b.ctx.Done():
						}
					}
				}
			}
		}
	})

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

// 获取浏览器
func (m *Manager) GetBrowser(id int64) (*Browser, error) {
	if bs, ok := m.browsers[id]; ok {
		return bs, nil
	}
	return nil, fmt.Errorf("浏览器 %d 不存在", id)
}
