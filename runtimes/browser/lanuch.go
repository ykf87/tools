package browser

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func (this *User) Lanuch(dataDir string) error {
	lanuch := launcher.New().
		Leakless(false).
		Bin(BROWSERFILE). // 使用 VirtualBrowser
		Headless(false).  // 是否显示窗口
		Set("user-data-dir", dataDir).
		Set("worker-id", fmt.Sprintf("%d", this.Id))

	wsurl, err := lanuch.Launch()
	if err != nil {
		return err
	}
	this.LanucherUrl = wsurl

	this.browser = rod.New().ControlURL(wsurl)
	if err := this.browser.Connect(); err != nil {
		this.browser.Close()
		return err
	}
	return nil
}

func (this *User) Page(urlStr string) error {
	//  this.browser.mustpa
	return nil
}
