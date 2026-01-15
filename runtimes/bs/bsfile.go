package bs

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"tools/runtimes/funcs"
)

// 生成浏览器目录和配置文件
func MakeBrowserConfig(id int64, lang, timezone, proxy string) (*BrowserConfigFile, error) {
	dir, err := GetBrowserConfigDir(id)
	if err != nil {
		return nil, err
	}

	bs := new(BrowserConfigFile)
	bs.Id = id

	var oldUser *BrowserConfigFile
	configFile := filepath.Join(dir, configFileName)
	if bt, err := os.ReadFile(configFile); err == nil {
		cfg := new(VirtualBrowserConfig)
		if Json.Unmarshal(bt, cfg) == nil && len(cfg.Users) > 0 {
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

	if proxy != "" {
		bs.SetProxy(proxy, "", "")
	} else {
		bs.Proxy.Mode = 0
	}

	bs.Screen.Mode = 0
	bs.SecChUa.Mode = 0

	bs.SecChUa.Value = append(bs.SecChUa.Value,
		SecChUaStruct{Brand: "Chromium", Version: 120},
		SecChUaStruct{Brand: "Not=A?Brand", Version: "99"},
	)

	if timezone == "" {
		timezone = "PRC"
	}
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

	cfg := new(VirtualBrowserConfig)
	cfg.Users = append(cfg.Users, bs)
	bt, err := Json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	if err := funcs.SaveFile(configFile, bytes.NewReader(bt)); err != nil {
		return nil, err
	}

	return bs, nil
}

// 暂时废弃
func (u *BrowserConfigFile) SetProxy(proxyurl, user, password string) {
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
