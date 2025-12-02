package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tools/runtimes/funcs"
	"tools/runtimes/requests"
)

const (
	LOGROOT      = "logs"                           // 日志目录
	CACHEROOT    = "cache"                          // 缓存目录
	DATAROOT     = "data"                           // 媒体文件路径
	WEBROOT      = ".web"                           // 网页端文件路径,开头是.的默认隐藏
	SYSROOT      = ".sys"                           // 系统存储的文件
	DBFILE       = SYSROOT + "/.db"                 // 数据库文件
	VERSION      = "1.0.0"                          // 字符串版本
	VERSIONCODE  = 100                              // 整数版本
	PROXYMINPORT = 100                              // 代理最小的端口号
	BROWSERCACHE = SYSROOT + "/browsers/cache"      // 浏览器缓存
	MEDIAROOT    = DATAROOT + "/media"              // 媒体文件路径
	SERVERDOMAIN = "http://127.0.0.1:20250/server/" // 服务端的地址
	SERVERWS     = "ws://127.0.0.1:20250/server/ws" // 服务端ws连接地址
)

var RuningRoot string

type mkdirStruct struct {
	DirName string      `json:"dir_name"`
	Mode    fs.FileMode `json:"mode"`
	IsHide  bool        `json:"is_hide"`
}

var WebFilesDownUrl = "http://127.0.0.1" // 网页下载地址

var Lang = "zh-cn"
var Timezone = "PRC"
var Currency = "CNY"
var CurrRate = 1.0
var Country = "cn"
var TimeFormat = "15:04:05"
var DateFormat = "2006-01-02"
var DateTimeFormat = "2006-01-02 15:04:05"
var ApiUrl = ""
var WebUrl = ""
var MediaUrl = ""

var Mkdirs = map[string]*mkdirStruct{
	"log": &mkdirStruct{
		DirName: LOGROOT,
		Mode:    os.ModePerm,
		IsHide:  false,
	},
	"cache": &mkdirStruct{
		DirName: LOGROOT,
		Mode:    os.ModePerm,
		IsHide:  false,
	},
	"node": &mkdirStruct{
		DirName: WEBROOT,
		Mode:    os.ModePerm,
		IsHide:  true,
	},
	"data": &mkdirStruct{
		DirName: DATAROOT,
		Mode:    os.ModePerm,
		IsHide:  false,
	},
	"sys": &mkdirStruct{
		DirName: SYSROOT,
		Mode:    os.ModePerm,
		IsHide:  true,
	},
	"browser": &mkdirStruct{
		DirName: BROWSERCACHE,
		Mode:    os.ModePerm,
		IsHide:  true,
	},
	"media": &mkdirStruct{
		DirName: MEDIAROOT,
		Mode:    os.ModePerm,
		IsHide:  false,
	},
}

func FullPath(pathName ...string) string {
	// 过滤掉空的 pathName，避免出现多余的 "/"
	cleaned := make([]string, 0, len(pathName))
	for _, p := range pathName {
		if p != "" {
			if strings.Contains(p, RuningRoot) {
				p = strings.TrimLeft(p, RuningRoot)
			}
			cleaned = append(cleaned, p)
		}
	}

	// Join 时自动处理多余的分隔符
	full := filepath.Join(append([]string{RuningRoot}, cleaned...)...)

	// 转换为绝对路径，保证一致性
	abs, err := filepath.Abs(full)
	if err != nil {
		return full // 出错就返回 Join 的结果
	}
	return abs
}

// 从远程获取版本信息
type Versions struct {
	Id          int64  `json:"id"`
	Code        string `json:"code"`
	CodeNum     int64  `json:"code_num"`
	Title       string `json:"title"`
	Desc        string `json:"desc"`
	Content     string `json:"content"`
	Os          string `json:"os"`
	Released    int    `json:"released"`
	Addtime     int64  `json:"addtime"`
	ReleaseTime int64  `json:"release_time"`
}
type VersionResp struct {
	Code int         `json:"code"`
	Data []*Versions `json:"data"`
	Msg  string      `json:"msg"`
}

var VersionResps *VersionResp

func GetVersions() *VersionResp {
	VersionResps = new(VersionResp)
	if r, err := requests.New(&requests.Config{Timeout: time.Second * 10}); err == nil {
		hd := funcs.ServerHeader(VERSION, VERSIONCODE)
		if str, err := r.Get(fmt.Sprint(SERVERDOMAIN, "versions"), hd); err == nil {
			rsp := new(VersionResp)
			if err := json.Unmarshal([]byte(str), rsp); err == nil {
				VersionResps = rsp
			}
		}
	}
	return VersionResps
}
