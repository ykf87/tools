package browser

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
)

var BROWSERPATH = ""
var BROWSERFILE = ""

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

func NewBrowser(lang string) *User {
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
	return bs
}

func (this *User) SetProxy(proxyurl, user, password string) {

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
