package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"tools/runtimes/eventbus"
	"tools/runtimes/funcs"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	json "github.com/json-iterator/go"
	// "gvisor.dev/gvisor/runsc/cmd"
)

func (this *User) LanuchCmd(dataDir string) error {
	userDataFile := filepath.Join(dataDir, configFileName)
	cc := new(VirtualBrowserConfig)

	if _, err := os.Stat(userDataFile); err == nil {
		os.Remove(userDataFile)
	}

	cc.Users = append(cc.Users, this)
	content, err := json.Marshal(cc)
	if err != nil {
		return err
	}
	if err := os.WriteFile(userDataFile, content, os.ModePerm); err != nil {
		return err
	}

	// cmd := new(funcs.Command)
	if this.ListenPort, err = funcs.FreePort(); err != nil {
		return err
	}
	binName := BROWSERFILE
	_, ccc, err := funcs.RunCommand(false, binName,
		"--remote-debugging-address=0.0.0.0",
		fmt.Sprintf("--remote-debugging-port=%d", this.ListenPort),
		"--flag-switches-begin",
		"--flag-switches-end",
		fmt.Sprintf("--user-data-dir=%s", dataDir),
		fmt.Sprintf("--worker-id=%d", this.Id),
		"--origin-trial-disabled-features=CanvasTextNg|WebAssemblyCustomDescriptors",
	)
	if err != nil {
		return err
	}
	this.Cmd = ccc

	go func() {
		this.Cmd.Wait()
		// fmt.Println("-0---- 浏览器关闭")
		this.Close()
	}()
	return nil
}

func (this *User) Lanuch(dataDir string) error {
	if err := this.LanuchCmd(dataDir); err != nil {
		return err
	}
	return nil
	userDataFile := filepath.Join(dataDir, configFileName)
	cc := new(VirtualBrowserConfig)

	if _, err := os.Stat(userDataFile); err == nil {
		os.Remove(userDataFile)
	}

	cc.Users = append(cc.Users, this)
	content, err := json.Marshal(cc)
	if err != nil {
		return err
	}
	if err := os.WriteFile(userDataFile, content, os.ModePerm); err != nil {
		return err
	}

	remotePort, err := funcs.FreePort()
	if err != nil {
		return err
	}

	binName := BROWSERFILE
	// binName = "C:/Program Files/Google/Chrome/Application/chrome.exe"

	lanuch := launcher.New().
		Leakless(false).
		Bin(binName).    // 使用 VirtualBrowser
		Headless(false). // 是否显示窗口
		Set("user-data-dir", dataDir).
		Set("worker-id", fmt.Sprintf("%d", this.Id)).
		Set("start-maximized").
		Set("no-sandbox").
		Set("disable-infobars").
		Set("user-agent", this.Ua.Value).
		// Set("disable-blink-features", "AutomationControlled").
		// Delete("headless").
		Set("load-extension", "").
		// Set("flag-switches-begin").
		// Set("flag-switches-end").
		Set("proxy-server", this.Proxy.Url).
		// Set("lang", this.UaLanguage.Value).
		Set("remote-debugging-port", fmt.Sprintf("%d", remotePort)).
		Set("window-size", fmt.Sprintf("%d,%d", this.Screen.Width, this.Screen.Height)).
		UserDataDir(dataDir).
		Delete("enable-automation").
		Delete("no-startup-window").
		// Delete("disable-background-networking").
		// Delete("disable-background-timer-throttling").
		// Delete("disable-backgrounding-occluded-windows").
		// Delete("disable-breakpad").
		// Delete("disable-client-side-phishing-detection").
		// Delete("disable-component-extensions-with-background-pages").
		// Delete("disable-default-apps").
		// Delete("disable-dev-shm-usage").
		// Delete("disable-features").
		// Delete("disable-hang-monitor").
		// Delete("disable-infobars").
		// Delete("disable-ipc-flooding-protection").
		// Delete("disable-popup-blocking").
		// Delete("disable-prompt-on-repost").
		// Delete("disable-renderer-backgrounding").
		// Delete("disable-site-isolation-trials").
		// Delete("disable-sync").
		// Delete("enable-features").
		// Delete("metrics-recording-only").
		// Delete("remote-debugging-port").
		// Delete("use-mock-keychain").
		// Delete("origin-trial-disabled-features").
		// Delete("force-color-profile").
		// Delete("load-extension").
		// Delete("no-first-run").
		Delete("no-sandbox")
		// Delete("start-maximized")
		// Set("remote-debugging-port", "9222")
		// Set("headless", "false")
		// fmt.Println(lanuch.FormatArgs(), "-----")
	if this.UaLanguage.Value != "" {
		lanuch.Set("lang", this.UaLanguage.Value)
	}
	wsurl := lanuch.MustLaunch()

	this.LanucherUrl = wsurl

	this.browser = rod.New().ControlURL(wsurl)
	if err := this.browser.Connect(); err != nil {
		this.browser.Close()
		return err
	}

	if this.Homepage.Value != "" {
		if err := this.Page(this.Homepage.Value, 0); err != nil {
			return err
		}
	} else {
		if err := this.Page("about:blank", 0); err != nil {
			return err
		}
	}

	go this.listens()

	return nil
}

func (this *User) setWindowSize(p *rod.Page, w, h *int) error {
	// 获取 windowId
	// res, err := proto.BrowserGetWindowForTarget{}.Call(p)
	// if err != nil {
	// 	return err
	// }
	// // 设置窗口 bounds
	// pbswb := proto.BrowserSetWindowBounds{
	// 	WindowID: res.WindowID,
	// 	Bounds: &proto.BrowserBounds{
	// 		WindowState: proto.BrowserWindowStateNormal,
	// 		Width:       w,
	// 		Height:      h,
	// 	},
	// }
	// if err := pbswb.Call(p); err != nil {
	// 	return err
	// }

	pesd := proto.EmulationSetDeviceMetricsOverride{
		Width:             *w,
		Height:            *h,
		DeviceScaleFactor: 1, // 或按你的 DPI 设置
		Mobile:            false,
	}
	if err := pesd.Call(p); err != nil {
		return err
	}

	p.MustEvalOnNewDocument(`() => {
		try {
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
			Object.defineProperty(navigator, 'languages', { get: () => ['en-US','en'] });
			Object.defineProperty(navigator, 'plugins', { get: () => [1,2,3] });
			window.resizeTo = function() {};
			window.resizeBy = function() {};
		} catch (e) {}
	}`)
	return nil
}

func (this *User) Close() error {
	if this.browser != nil {
		this.browser.Close()
	}
	if this.Cmd != nil {
		this.Cmd.Process.Kill()
	}

	eventbus.Bus.Publish("browser-close", this)
	return nil
}

func (this *User) Page(urlStr string, tagIndex int) error {
	pages, _ := this.browser.Pages()
	if tagIndex < 0 {
		tagIndex = 0
	}
	pl := len(pages)
	var pg *rod.Page
	if tagIndex <= (pl - 1) {
		pg = pages[tagIndex]
		pg.Navigate(urlStr)
	} else {
		pg = this.browser.MustPage(urlStr)
	}
	this.setWindowSize(pg, &this.Screen.Width, &this.Screen.Height)
	return nil
}

func (this *User) listens() {
	for e := range this.browser.Event() {
		if e.Method == "Target.targetCreated" || e.Method == "Target.targetDestroyed" {
			pages, _ := this.browser.Pages()
			fmt.Println("当前标签页数量:", len(pages))
			for _, v := range pages {
				this.setWindowSize(v, &this.Screen.Width, &this.Screen.Height)
			}
		}
		// 什么都不做，只是消费事件
	}
	this.Close()
}
