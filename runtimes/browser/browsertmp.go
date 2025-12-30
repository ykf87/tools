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
	"tools/runtimes/db/messages"
	"tools/runtimes/db/notifies"
	"tools/runtimes/funcs"
	"tools/runtimes/services"

	json "github.com/json-iterator/go"
)

// var BROWSERPATH = ""
// var BROWSERFILE = ""
// var BROWSERDOWNLOADED bool

// var Running sync.Map

var (
	BROWSERPATH string
	BROWSERFILE string

	Running sync.Map

	browserOnce sync.Once
	browserErr  error
)

func init() {
	BROWSERPATH = filepath.Join(config.SYSROOT, "browser")
	EnsureBrowser()

	// needdownload := true
	// BROWSERPATH = filepath.Join(config.SYSROOT, "browser")
	// var filename string
	// switch runtime.GOOS {
	// case "darwin":
	// 	filename = "VirtualBrowser.dmg"
	// default:
	// 	filename = "VirtualBrowser.exe"
	// }
	// if _, err := os.Stat(BROWSERPATH); err == nil {
	// 	bf := config.FullPath(BROWSERPATH, filename)
	// 	if _, err := os.Stat(bf); err != nil {
	// 		needdownload = true
	// 	} else {
	// 		needdownload = false
	// 		BROWSERFILE = bf
	// 	}
	// }
	// if needdownload == true {
	// 	fmt.Println("下载 browser...")
	// 	go DownBrowserFromServer(BROWSERPATH)
	// }
}

/* =========================
   Browser 准备
========================= */

func EnsureBrowser() error {
	browserOnce.Do(func() {
		browserErr = prepareBrowser()
	})
	return browserErr
}

func prepareBrowser() error {
	var filename string
	switch runtime.GOOS {
	case "darwin":
		filename = "VirtualBrowser.dmg"
	default:
		filename = "VirtualBrowser.exe"
	}

	bf := config.FullPath(BROWSERPATH, filename)
	if _, err := os.Stat(bf); err == nil {
		BROWSERFILE = bf
		return nil
	}

	fmt.Println("下载 browser...")
	if err := DownBrowserFromServer(BROWSERPATH); err != nil {
		return err
	}

	if _, err := os.Stat(bf); err != nil {
		return fmt.Errorf("browser 文件不存在: %s", bf)
	}

	BROWSERFILE = bf
	return nil
}

// 从服务端下载指纹浏览器
func DownBrowserFromServer(saveto string) error {
	// time.Sleep(time.Second * 3)
	// nty := &notifies.Notify{
	// 	Type:        "error",
	// 	Title:       "指纹浏览器",
	// 	Description: "指纹浏览器下载失败",
	// 	Url:         config.ApiUrl + "/browser/download",
	// 	Btn:         "重新下载",
	// 	Method:      "get",
	// }

	// serverBrowserName := "browser.zip"
	// downurl := fmt.Sprint(config.SERVERDOMAIN, "down?file=browser/"+serverBrowserName)
	// saveFile := filepath.Join(saveto, serverBrowserName)
	// if err := services.ServerDownload(downurl, saveFile, nil, func(perc float64, downloaded, total int64) {
	// 	fmt.Printf("\r下载中：%.2f%% (%s/%s)", perc, funcs.FormatFileSize(downloaded), funcs.FormatFileSize(total))
	// }); err != nil {
	// 	nty.Meta = time.Now().Format("2006-01-02 15:04:05")
	// 	nty.Content = err.Error()
	// 	nty.Send()
	// 	BROWSERDOWNLOADED = false
	// 	return err
	// }
	// fmt.Println("\n下载完成,开始解压......")
	// // 解压文件
	// if err := funcs.Unzip(saveFile, saveto); err != nil {
	// 	nty.Meta = time.Now().Format("2006-01-02 15:04:05")
	// 	nty.Content = err.Error()
	// 	nty.Send()
	// 	BROWSERDOWNLOADED = false
	// 	return err
	// }
	// fmt.Println("解压完成!")
	// os.Remove(saveFile)

	// msg := &messages.Message{
	// 	Type:    "success",
	// 	Content: "指纹浏览器下载成功!",
	// }
	// msg.Send()
	// BROWSERDOWNLOADED = true
	// return nil
	//

	nty := &notifies.Notify{
		Type:        "error",
		Title:       "指纹浏览器",
		Description: "指纹浏览器下载失败",
		Url:         config.ApiUrl + "/browser/download",
		Btn:         "重新下载",
		Method:      "get",
	}

	serverBrowserName := "browser.zip"
	downurl := fmt.Sprint(config.SERVERDOMAIN, "down?file=browser/"+serverBrowserName)
	saveFile := filepath.Join(saveto, serverBrowserName)

	if err := services.ServerDownload(downurl, saveFile, nil, func(perc float64, downloaded, total int64) {
		fmt.Printf("\r下载中：%.2f%% (%s/%s)", perc,
			funcs.FormatFileSize(downloaded),
			funcs.FormatFileSize(total))
	}); err != nil {
		nty.Meta = time.Now().Format("2006-01-02 15:04:05")
		nty.Content = err.Error()
		nty.Send()
		return err
	}

	fmt.Println("\n下载完成,开始解压......")

	if err := funcs.Unzip(saveFile, saveto); err != nil {
		nty.Meta = time.Now().Format("2006-01-02 15:04:05")
		nty.Content = err.Error()
		nty.Send()
		return err
	}

	_ = os.Remove(saveFile)

	msg := &messages.Message{
		Type:    "success",
		Content: "指纹浏览器下载成功!",
	}
	msg.Send()

	return nil
}

func NewBrowser(lang, timezone string, id int64) *User {
	// if temu, ok := Running.Load(id); ok {
	// 	u, ok := temu.(*User)
	// 	if !ok {
	// 		Running.Delete(id)
	// 	} else {
	// 		return u
	// 	}
	// }
	// bs := new(User)
	// bs.Id = id

	// var oldUser *User
	// configFile := filepath.Join(bs.WorkDir(), configFileName)
	// if _, err := os.Stat(configFile); err == nil {
	// 	if cbt, err := os.ReadFile(configFile); err == nil {
	// 		odc := new(VirtualBrowserConfig)
	// 		if err := json.Unmarshal(cbt, odc); err == nil {
	// 			oldUser = odc.Users[0]
	// 		}
	// 	}
	// }

	// bs.AudioContext = new(AudioContextStruct)
	// bs.Canvas = new(CanvasStruct)
	// bs.DeviceName = new(DeviceNameStruct)
	// bs.Proxy = new(PortStruct)
	// bs.Location = new(LocationStruct)
	// bs.Webgl = new(WebglStruct)
	// bs.WebglImg = new(WebglImgStruct)
	// bs.ClientRects = new(ClientRectsStruct)
	// bs.ChromeVersion = "默认"
	// bs.Cpu.Mode = 1
	// bs.Dnt.Mode = 1
	// bs.Dnt.Value = 0
	// bs.Fonts.Mode = 0
	// bs.Gpu.Mode = 1
	// bs.Group = "Default"
	// bs.Location.Mode = 2
	// bs.Mac.Mode = 1
	// bs.Media.Mode = 1
	// bs.Memory.Mode = 1
	// bs.Name = "Default Name"
	// bs.Os = runtime.GOOS
	// bs.PortScan.Mode = 1
	// bs.Proxy.Mode = 0
	// bs.Screen.Mode = 0

	// bs.SecChUa.Mode = 0
	// bs.SecChUa.Value = append(bs.SecChUa.Value, SecChUaStruct{
	// 	Brand:   "Chromium",
	// 	Version: 120,
	// }, SecChUaStruct{
	// 	Brand:   "Not=A?Brand",
	// 	Version: "99",
	// })
	// bs.SetTimezone(timezone)
	// bs.Ua.Mode = 0
	// bs.Ua.Value = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	// bs.UaFullVersion.Mode = 0
	// bs.UaFullVersion.Value = "120.0.6099.291"

	// if lang != "" {
	// 	bs.UaLanguage.Mode = 1
	// 	if strings.Contains(lang, "-") == false {
	// 		for k, _ := range LangMap {
	// 			if strings.Contains(k, lang) == true {
	// 				bs.UaLanguage.Language = k
	// 				bs.UaLanguage.Value = fmt.Sprintf("%s,%s", k, lang)
	// 				break
	// 			}
	// 		}
	// 	} else {
	// 		bs.UaLanguage.Language = lang
	// 		bs.UaLanguage.Value = lang
	// 	}
	// }
	// bs.Webrtc.Mode = 0
	// bs.SetHomePage("about:blank")

	// if oldUser != nil {
	// 	bs.AudioContext = oldUser.AudioContext
	// 	bs.Canvas = oldUser.Canvas
	// 	bs.ClientRects = oldUser.ClientRects
	// 	bs.DeviceName = oldUser.DeviceName
	// 	bs.Webgl = oldUser.Webgl
	// 	bs.WebglImg = oldUser.WebglImg
	// 	bs.Gpu = oldUser.Gpu
	// 	bs.Location.Enable = oldUser.Location.Enable
	// 	bs.Location.Precision = oldUser.Location.Precision
	// 	bs.Mac.Value = oldUser.Mac.Value
	// 	bs.Memory.Value = oldUser.Memory.Value
	// 	bs.Timestamp = oldUser.Timestamp
	// 	bs.Cpu.Value = oldUser.Cpu.Value
	// } else {
	// 	bs.AudioContext.Random()
	// 	bs.Canvas.Random()
	// 	bs.ClientRects.Random()
	// 	bs.DeviceName.Random()
	// 	bs.Webgl.Random()
	// 	bs.WebglImg.Random()
	// 	bs.Gpu.Value = rand.Intn(len([]int{1, 2, 4, 8, 16}))
	// 	bs.Location.Enable = rand.Intn(len([]int{2, 3}))
	// 	bs.Location.Precision = rand.Intn(4000-200+1) + 200
	// 	bs.Mac.Value = strings.ReplaceAll(funcs.RandomMAC(""), ":", "-")
	// 	bs.Memory.Value = rand.Intn(len([]int{4, 8, 16, 32, 64, 128}))
	// 	bs.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	// 	bs.Cpu.Value = rand.Intn(len([]int{1, 2, 4, 8, 12}))
	// }
	// return bs
	//
	if u, ok := Running.Load(id); ok {
		if usr, ok := u.(*User); ok {
			return usr
		}
		Running.Delete(id)
	}

	bs := new(User)
	bs.Id = id

	var oldUser *User
	configFile := filepath.Join(bs.WorkDir(), configFileName)
	if bt, err := os.ReadFile(configFile); err == nil {
		cfg := new(VirtualBrowserConfig)
		if json.Unmarshal(bt, cfg) == nil && len(cfg.Users) > 0 {
			oldUser = cfg.Users[0]
		}
	}

	/* 默认配置 */
	bs.AudioContext = new(AudioContextStruct)
	bs.Canvas = new(CanvasStruct)
	bs.DeviceName = new(DeviceNameStruct)
	bs.Proxy = new(PortStruct)
	bs.Location = new(LocationStruct)
	bs.Webgl = new(WebglStruct)
	bs.WebglImg = new(WebglImgStruct)
	bs.ClientRects = new(ClientRectsStruct)

	bs.ChromeVersion = "默认"
	bs.Group = "Default"
	bs.Name = "Default Name"
	bs.Os = runtime.GOOS

	bs.Cpu.Mode = 1
	bs.Dnt.Mode = 1
	bs.Dnt.Value = 0
	bs.Fonts.Mode = 0
	bs.Gpu.Mode = 1
	bs.Location.Mode = 2
	bs.Mac.Mode = 1
	bs.Media.Mode = 1
	bs.Memory.Mode = 1
	bs.PortScan.Mode = 1
	bs.Proxy.Mode = 0
	bs.Screen.Mode = 0
	bs.SecChUa.Mode = 0

	bs.SecChUa.Value = append(bs.SecChUa.Value,
		SecChUaStruct{Brand: "Chromium", Version: 120},
		SecChUaStruct{Brand: "Not=A?Brand", Version: "99"},
	)

	bs.SetTimezone(timezone)
	bs.Ua.Mode = 0
	bs.Ua.Value = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	bs.UaFullVersion.Mode = 0
	bs.UaFullVersion.Value = "120.0.6099.291"

	if lang != "" {
		bs.UaLanguage.Mode = 1
		if !strings.Contains(lang, "-") {
			for k := range LangMap {
				if strings.Contains(k, lang) {
					bs.UaLanguage.Language = k
					bs.UaLanguage.Value = fmt.Sprintf("%s,%s", k, lang)
					break
				}
			}
		} else {
			bs.UaLanguage.Language = lang
			bs.UaLanguage.Value = lang
		}
	}

	bs.Webrtc.Mode = 0
	bs.SetHomePage("about:blank")

	/* 恢复 or 随机 */
	if oldUser != nil {
		*bs = *oldUser
		bs.Id = id
	} else {
		bs.AudioContext.Random()
		bs.Canvas.Random()
		bs.ClientRects.Random()
		bs.DeviceName.Random()
		bs.Webgl.Random()
		bs.WebglImg.Random()

		bs.Gpu.Value = []int{1, 2, 4, 8, 16}[rand.Intn(5)]
		bs.Memory.Value = []int{4, 8, 16, 32, 64, 128}[rand.Intn(6)]
		bs.Cpu.Value = []int{1, 2, 4, 8, 12}[rand.Intn(5)]
		bs.Location.Enable = []int{2, 3}[rand.Intn(2)]
		bs.Location.Precision = rand.Intn(3801) + 200
		bs.Mac.Value = strings.ReplaceAll(funcs.RandomMAC(""), ":", "-")
		bs.Timestamp = time.Now().UnixMilli()
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
	// if u, ok := Running.Load(this.Id); ok {
	// 	usr := u.(*User)
	// 	return usr, nil
	// }
	// // cc := new(VirtualBrowserConfig)
	// // cc.Users = append(cc.Users, this)
	// // bt, err := json.Marshal(cc)
	// // if err != nil {
	// // 	return nil, err
	// // }

	// // wk := filepath.Join(config.BROWSERCACHE, fmt.Sprintf("%d", this.Id))
	// wk := this.WorkDir()
	// if _, err := os.Stat(wk); err != nil {
	// 	if err := os.MkdirAll(wk, os.ModePerm); err != nil {
	// 		return nil, err
	// 	}
	// }

	// // if err = os.WriteFile(wk, bt, 0644); err != nil {
	// // 	return nil, err
	// // }

	// this.Lanuch(wk)

	// Running.Store(this.Id, this)
	// return this, nil

	if err := EnsureBrowser(); err != nil {
		return nil, err
	}

	if v, loaded := Running.LoadOrStore(this.Id, this); loaded {
		return v.(*User), nil
	}

	wk := this.WorkDir()
	if err := os.MkdirAll(wk, 0755); err != nil {
		Running.Delete(this.Id)
		return nil, err
	}

	if err := this.Lanuch(wk); err != nil {
		Running.Delete(this.Id)
		return nil, err
	}

	return this, nil
}

func (u *User) SetProxy(proxyurl, user, password string) {
	if proxyurl == "" {
		return
	}

	pu, err := url.Parse(proxyurl)
	if err != nil {
		return
	}

	host, port, _ := net.SplitHostPort(pu.Host)
	if host == "" {
		host = pu.Host
	}
	if port == "" {
		if pu.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	u.Proxy.Mode = 2
	u.Proxy.Url = proxyurl
	u.Proxy.User = user
	u.Proxy.Pass = password
	u.Proxy.Host = host
	u.Proxy.Port = port
	u.Proxy.Protocol = strings.ToUpper(pu.Scheme)
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
