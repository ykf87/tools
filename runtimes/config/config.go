package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"tools/runtimes/funcs"

	jsoniter "github.com/json-iterator/go"
)

var Json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	LOGROOT      = "logs"                           // 日志目录
	CACHEROOT    = "cache"                          // 缓存目录
	DATAROOT     = "data"                           // 媒体文件路径
	WEBROOT      = ".web"                           // 网页端文件路径,开头是.的默认隐藏
	SYSROOT      = ".sys"                           // 系统存储的文件
	DBFILE       = ".db"                            // 数据库文件
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
var MediaUrl = ""
var MainStatus = 1
var MainStatusMsg string
var ApiPort int
var WebPort int
var ApiUrl = ""
var WebUrl = ""
var BrowserReportJs = ""
var AdminWidthAndHeight sync.Map

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
	"mini": &mkdirStruct{
		DirName: ".mini",
		Mode:    os.ModePerm,
		IsHide:  true,
	},
}

func init() {
	for _, v := range Mkdirs {
		full := FullPath(v.DirName)
		if _, err := os.Stat(full); err != nil {
			if err := os.MkdirAll(full, v.Mode); err != nil {
				panic(err)
			}
			if v.IsHide == true {
				funcs.HiddenDir(full)
			}
		}
	}

	// 还要autojs的上报封装,最好能从服务端获取最新的
	BrowserReportJs = `class Callback{constructor(options={}){this.app=options.app||'unknown';this.env=options.env||'dev';this.enable=options.enable!==false;this.version="1.3.26"}report(payload){if(!this.enable)return;const body={app:this.app,env:this.env,version:this.version,time:Date.now(),...payload};console.log(JSON.stringify(body))}success(msg,data=null){this.report({type:'success',msg,data})}fail(msg,error=null){this.report({type:'fail',msg,data:error})}notify(msg,data=null){this.report({type:'notify',msg,data})}upload(clickNode,fileNode,files){this.report({type:'upload',msg:'',data:{node:clickNode,upnode:fileNode,files:files}})}input(node,text){this.report({type:'input',msg:'',data:{node:clickNode,text:text}})}invoke(name,params=null){this.report({type:'invoke',msg:name,data:params})}state(name,value){this.report({type:'state',msg:name,data:value})}event(name,data=null){this.report({type:'event',msg:name,data})}metric(name,value,unit=''){this.report({type:'metric',msg:name,data:{value,unit}})}}`
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

type Version struct {
	Id          int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Code        string `json:"code" gorm:"uniqueIndex;not null"`    // 版本编码
	CodeNum     int64  `json:"code_num" gorm:"index;default:0"`     // 版本纯数字
	Title       string `json:"title" gorm:"index;default:null"`     // 标题
	Desc        string `json:"desc" gorm:"default:null"`            // 版本简介
	Content     string `json:"content" gorm:"default:null"`         // 版本详情
	Os          string `json:"os" gorm:"index;default:null"`        // 支持的系统
	Released    int    `json:"released" gorm:"index;default:0"`     // 是否发布
	Addtime     int64  `json:"addtime" gorm:"index;default:0"`      // 添加时间
	ReleaseTime int64  `json:"release_time" gorm:"index;default:0"` // 发布时间
}

var Versions []*Version

// 从服务端获取的大于等于当前版本的版本信息列表
func VersionsFromServer(vs string) error {
	if vs != "" {
		if err := Json.Unmarshal([]byte(vs), &Versions); err != nil {
			return err
		}
	}
	return nil
}
