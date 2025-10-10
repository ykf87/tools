package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"tools/runtimes/eventbus"
	"tools/runtimes/funcs"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	json "github.com/json-iterator/go"
)

func (this *User) Lanuch(dataDir string) error {
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

	lanuch := launcher.New().
		Leakless(false).
		Bin(BROWSERFILE). // 使用 VirtualBrowser
		Headless(false).  // 是否显示窗口
		Set("user-data-dir", dataDir).
		Set("worker-id", fmt.Sprintf("%d", this.Id)).
		Set("start-maximized").
		Set("no-sandbox").
		Set("disable-infobars").
		Set("user-agent", this.Ua.Value).
		Delete("enable-automation").
		Delete("headless").
		Set("load-extension", "").
		Set("flag-switches-begin").
		Set("flag-switches-end").
		Set("proxy-server", this.Proxy.Url).
		Set("lang", this.UaLanguage.Value).
		Set("remote-debugging-port", fmt.Sprintf("%d", remotePort)).
		Set("window-size", fmt.Sprintf("%d,%d", this.Screen.Width, this.Screen.Height)).
		UserDataDir(dataDir).
		Delete("no-startup-window").
		Delete("disable-background-networking").
		Delete("disable-background-timer-throttling").
		Delete("disable-backgrounding-occluded-windows").
		Delete("disable-breakpad").Delete("disable-client-side-phishing-detection").
		Delete("disable-component-extensions-with-background-pages").
		Delete("disable-default-apps").
		Delete("disable-dev-shm-usage").
		Delete("disable-features").
		Delete("disable-hang-monitor").
		Delete("disable-infobars").
		Delete("disable-ipc-flooding-protection").
		Delete("disable-popup-blocking").
		Delete("disable-prompt-on-repost").
		Delete("disable-renderer-backgrounding").
		Delete("disable-site-isolation-trials").
		Delete("disable-sync").
		Delete("enable-features").
		Delete("metrics-recording-only").
		// Delete("remote-debugging-port").
		Delete("use-mock-keychain").
		Delete("origin-trial-disabled-features").
		Delete("force-color-profile").
		Delete("load-extension").
		Delete("no-first-run").
		Delete("no-sandbox").
		Delete("start-maximized")
		// Set("remote-debugging-port", "9222")
		// Set("headless", "false")
		// fmt.Println(lanuch.FormatArgs(), "-----")
	wsurl := lanuch.MustLaunch()

	this.LanucherUrl = wsurl

	this.browser = rod.New().ControlURL(wsurl)
	if err := this.browser.Connect(); err != nil {
		this.browser.Close()
		return err
	}

	if this.Homepage.Value != "" {
		this.Page(this.Homepage.Value)
	}

	go this.listens()

	return nil
}

func (this *User) Close() error {
	if this.browser != nil {
		this.browser.Close()
	}

	eventbus.Bus.Publish("browser-close", this)
	return nil
}

func (this *User) Page(urlStr string) error {
	this.browser.MustPage(urlStr)
	return nil
}

func (this *User) listens() {
	for e := range this.browser.Event() {
		if e.Method == "Target.targetCreated" || e.Method == "Target.targetDestroyed" {
			pages, _ := this.browser.Pages()
			fmt.Println("当前标签页数量:", len(pages))
		}
		// 什么都不做，只是消费事件
	}
	this.Close()
}
