package config

import (
	"io/fs"
	"os"
	"path/filepath"
)

const (
	LOGROOT      = "logs"           // 日志目录
	CACHEROOT    = "cache"          // 缓存目录
	DATAROOT     = "data"           // 媒体文件路径
	WEBROOT      = ".web"           // 网页端文件路径,开头是.的默认隐藏
	SYSROOT      = ".sys"           // 系统存储的文件
	DBFILE       = SYSROOT + "/.db" // 数据库文件
	VERSION      = "1.0.0"          // 字符串版本
	VERSIONCODE  = 100              // 整数版本
	PROXYMINPORT = 100              // 代理最小的端口号
)

type mkdirStruct struct {
	DirName string      `json:"dir_name"`
	Mode    fs.FileMode `json:"mode"`
	IsHide  bool        `json:"is_hide"`
}

var WebFilesDownUrl = "https://127.0.0.1" // 网页下载地址

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

var RuningRoot string
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
}

func FullPath(pathName ...string) string {
	// 过滤掉空的 pathName，避免出现多余的 "/"
	cleaned := make([]string, 0, len(pathName))
	for _, p := range pathName {
		if p != "" {
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
