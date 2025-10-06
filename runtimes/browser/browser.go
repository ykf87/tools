package browser

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"

	json "github.com/json-iterator/go"
)

var BROWSERPATH = ""
var BROWSERFILE = ""

var Running sync.Map

func init() {
	needdownload := false
	BROWSERPATH = filepath.Join(config.SYSROOT, "browser")
	if _, err := os.Stat(BROWSERPATH); err == nil {
		bf := filepath.Join(BROWSERPATH, "VirtualBrowser.exe")
		if _, err := os.Stat(bf); err != nil {
			needdownload = true
		} else {
			BROWSERFILE = bf
		}
	}
	if needdownload == true {
		fmt.Println("需下载 browser...")
		panic("-------")
	}
}

func NewBrowser(lang, timezone string) *User {
	bs := new(User)
	bs.AudioContext = new(AudioContextStruct)
	bs.AudioContext.Random()
	bs.Canvas = new(CanvasStruct)
	bs.Canvas.Random()
	bs.ClientRects = new(ClientRectsStruct)
	bs.ClientRects.Random()
	bs.ChromeVersion = "默认"
	bs.DeviceName = new(DeviceNameStruct)
	bs.DeviceName.Random()
	bs.Cpu.Mode = 1
	bs.Cpu.Value = 12
	bs.Dnt.Mode = 1
	bs.Dnt.Value = 0
	bs.Fonts.Mode = 0
	bs.Gpu.Mode = 1
	bs.Gpu.Value = rand.Intn(len([]int{1, 2, 4, 8, 16}))
	bs.Group = "Default"
	bs.Location.Mode = 2
	bs.Location.Enable = rand.Intn(len([]int{2, 3}))
	bs.Location.Precision = rand.Intn(4000-200+1) + 200
	bs.Mac.Mode = 1
	bs.Mac.Value = strings.ReplaceAll(funcs.RandomMAC(""), ":", "-")
	bs.Media.Mode = 1
	bs.Memory.Mode = 1
	bs.Memory.Value = rand.Intn(len([]int{4, 8, 16, 32, 64, 128}))
	bs.Name = "Default Name"
	bs.Os = runtime.GOOS
	bs.PortScan.Mode = 1
	bs.Proxy = new(PortStruct)
	bs.Proxy.Mode = 0
	bs.Screen.Mode = 0

	bs.SecChUa.Mode = 0
	bs.SecChUa.Value = append(bs.SecChUa.Value, SecChUaStruct{
		Brand:   "Chromium",
		Version: 120,
	}, SecChUaStruct{
		Brand:   "Not=A?Brand",
		Version: "99",
	})

	bs.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)

	bs.SetTimezone(timezone)
	bs.Ua.Mode = 0
	bs.Ua.Value = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	bs.UaFullVersion.Mode = 0
	bs.UaFullVersion.Value = "120.0.6099.291"

	bs.UaLanguage.Mode = 1
	bs.UaLanguage.Value = lang

	bs.Webgl = new(WebglStruct)
	bs.Webgl.Random()

	bs.WebglImg = new(WebglImgStruct)
	bs.WebglImg.Random()
	bs.Webrtc.Mode = 0
	return bs
}

type VirtualBrowserConfig struct {
	Users []*User `json:"users"`
}

func (this *User) Run(worker int64) (*User, error) {
	if u, ok := Running.Load(worker); ok {
		usr := u.(*User)
		return usr, nil
	}
	cc := new(VirtualBrowserConfig)
	cc.Users = append(cc.Users, this)
	bt, err := json.Marshal(cc)
	if err != nil {
		return nil, err
	}

	this.Id = worker

	wk := filepath.Join(config.BROWSERCACHE, fmt.Sprintf("%d", this.Id))
	if _, err := os.Stat(wk); err != nil {
		if err := os.MkdirAll(wk, os.ModePerm); err != nil {
			return nil, err
		}
	}

	if err = os.WriteFile(wk, bt, 0644); err != nil {
		return nil, err
	}

	this.Lanuch(wk)

	Running.Store(worker, this)
	return this, nil
}

func (this *User) SetProxy(proxyurl, user, password string) {
	if proxyurl != "" {
		this.Proxy.Mode = 2
		this.Proxy.User = user
		this.Proxy.Pass = password
	}
}

func (this *User) SetProxyApi(apiUrl string) {

}

func (this *User) SetCookie(cookie string) {
	// this.Cookie
}

func (this *User) SetScreen(width, height int) {
	this.Screen.Width = width
	this.Screen.Height = height
	this.Screen.Mode = 1
	this.Screen.Value = fmt.Sprintf("%d x %d", width, height)
}

func (this *User) SetTimezone(timezone string) {
	loc, err := time.LoadLocation(timezone)

	if err != nil {
		return
	}

	if this.TimeZone == nil {
		this.TimeZone = new(TimezoneStruct)
	}
	_, offset := time.Now().In(loc).Zone()
	hours := offset / 3600
	minutes := (offset % 3600) / 60

	sign := "+"
	if hours < 0 || minutes < 0 {
		sign = "-"
	}
	zone := fmt.Sprintf("UTC%s%02d:%02d", sign, abs(hours), abs(minutes))

	this.TimeZone.Locale = ""
	this.TimeZone.Mode = 2
	this.TimeZone.Name = this.TimeZone.GetName(timezone)
	this.TimeZone.Utc = timezone
	this.TimeZone.Value = 8
	this.TimeZone.Zone = zone
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
