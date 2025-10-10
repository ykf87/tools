package browser

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/eventbus"
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
		bf := config.FullPath(BROWSERPATH, "VirtualBrowser.exe")
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

	eventbus.Bus.Subscribe("browser-close", func(dt any) {
		if bu, ok := dt.(*User); ok {
			Running.Delete(bu.Id)
		}
	})
}

func NewBrowser(lang, timezone string, id int64) *User {
	if temu, ok := Running.Load(id); ok {
		u, ok := temu.(*User)
		if !ok {
			Running.Delete(id)
		} else {
			return u
		}
	}
	bs := new(User)
	bs.Id = id

	var oldUser *User
	configFile := filepath.Join(bs.WorkDir(), configFileName)
	if _, err := os.Stat(configFile); err == nil {
		if cbt, err := os.ReadFile(configFile); err == nil {
			odc := new(VirtualBrowserConfig)
			if err := json.Unmarshal(cbt, odc); err == nil {
				oldUser = odc.Users[0]
			}
		}
	}

	bs.AudioContext = new(AudioContextStruct)
	bs.Canvas = new(CanvasStruct)
	bs.DeviceName = new(DeviceNameStruct)
	bs.Proxy = new(PortStruct)
	bs.Location = new(LocationStruct)
	bs.Webgl = new(WebglStruct)
	bs.WebglImg = new(WebglImgStruct)
	bs.ClientRects = new(ClientRectsStruct)
	bs.ChromeVersion = "默认"
	bs.Cpu.Mode = 1
	bs.Dnt.Mode = 1
	bs.Dnt.Value = 0
	bs.Fonts.Mode = 0
	bs.Gpu.Mode = 1
	bs.Group = "Default"
	bs.Location.Mode = 2
	bs.Mac.Mode = 1
	bs.Media.Mode = 1
	bs.Memory.Mode = 1
	bs.Name = "Default Name"
	bs.Os = runtime.GOOS
	bs.PortScan.Mode = 1
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
	bs.SetTimezone(timezone)
	bs.Ua.Mode = 0
	bs.Ua.Value = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	bs.UaFullVersion.Mode = 0
	bs.UaFullVersion.Value = "120.0.6099.291"

	bs.UaLanguage.Mode = 1
	bs.UaLanguage.Value = lang
	bs.Webrtc.Mode = 0
	bs.SetHomePage("about:blank")

	if oldUser != nil {
		bs.AudioContext = oldUser.AudioContext
		bs.Canvas = oldUser.Canvas
		bs.ClientRects = oldUser.ClientRects
		bs.DeviceName = oldUser.DeviceName
		bs.Webgl = oldUser.Webgl
		bs.WebglImg = oldUser.WebglImg
		bs.Gpu = oldUser.Gpu
		bs.Location.Enable = oldUser.Location.Enable
		bs.Location.Precision = oldUser.Location.Precision
		bs.Mac.Value = oldUser.Mac.Value
		bs.Memory.Value = oldUser.Memory.Value
		bs.Timestamp = oldUser.Timestamp
		bs.Cpu.Value = oldUser.Cpu.Value
	} else {
		bs.AudioContext.Random()
		bs.Canvas.Random()
		bs.ClientRects.Random()
		bs.DeviceName.Random()
		bs.Webgl.Random()
		bs.WebglImg.Random()
		bs.Gpu.Value = rand.Intn(len([]int{1, 2, 4, 8, 16}))
		bs.Location.Enable = rand.Intn(len([]int{2, 3}))
		bs.Location.Precision = rand.Intn(4000-200+1) + 200
		bs.Mac.Value = strings.ReplaceAll(funcs.RandomMAC(""), ":", "-")
		bs.Memory.Value = rand.Intn(len([]int{4, 8, 16, 32, 64, 128}))
		bs.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		bs.Cpu.Value = rand.Intn(len([]int{1, 2, 4, 8, 12}))
	}
	return bs
}

type VirtualBrowserConfig struct {
	Users []*User `json:"users"`
}

func (this *User) WorkDir() string {
	return config.FullPath(config.BROWSERCACHE, fmt.Sprintf("%d", this.Id))
}

func (this *User) SetHomePage(url string) {
	this.Homepage.Mode = 1
	this.Homepage.Value = url
}

func (this *User) Run() (*User, error) {
	if u, ok := Running.Load(this.Id); ok {
		usr := u.(*User)
		return usr, nil
	}
	// cc := new(VirtualBrowserConfig)
	// cc.Users = append(cc.Users, this)
	// bt, err := json.Marshal(cc)
	// if err != nil {
	// 	return nil, err
	// }

	// wk := filepath.Join(config.BROWSERCACHE, fmt.Sprintf("%d", this.Id))
	wk := this.WorkDir()
	if _, err := os.Stat(wk); err != nil {
		if err := os.MkdirAll(wk, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// if err = os.WriteFile(wk, bt, 0644); err != nil {
	// 	return nil, err
	// }

	this.Lanuch(wk)

	Running.Store(this.Id, this)
	return this, nil
}

func (this *User) SetProxy(proxyurl, user, password string) {
	if proxyurl != "" {
		this.Proxy.Mode = 2
		this.Proxy.User = user
		this.Proxy.Pass = password
		this.Proxy.Url = proxyurl

		u, err := url.Parse(proxyurl)
		if err != nil {
			return
		}

		host := ""
		port := ""

		// 解析主机和端口
		if strings.Contains(u.Host, ":") {
			h, p, err := net.SplitHostPort(u.Host)
			if err == nil {
				host = h
				port = p
			} else {
				// 可能是 IPv6 或者没有端口
				host = u.Host
			}
		} else {
			host = u.Host
		}

		// 如果没写端口，自动补默认值
		if port == "" {
			if u.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		}
		this.Proxy.Port = port
		this.Proxy.Host = host
		this.Proxy.Protocol = "HTTP"
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

func Flush() {
	Running.Range(func(k, v any) bool {
		if bb, ok := v.(*User); ok {
			bb.Close()
		}
		Running.Delete(k)
		return true
	})
}
