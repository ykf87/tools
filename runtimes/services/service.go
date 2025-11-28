package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/requests"

	"github.com/tidwall/gjson"
)

// 从远程获取版本信息
type Versions struct{
	Id int64 `json:"id"`
	Code string `json:"code"`
	CodeNum int64 `json:"code_num"`
	Title string `json:"title"`
	Desc string `json:"desc"`
	Content string `json:"content"`
	Os string `json:"os"`
	Released int `json:"released"`
	Addtime int64 `json:"addtime"`
	ReleaseTime int64 `json:"release_time"`
}
type VersionResp struct{
	Code int `json:"code"`
	Data []*Versions `json:"data"`
	Msg string `json:"msg"`
}
var VersionResps *VersionResp


// 从服务端获取版本信息
func GetVersions() *VersionResp{
	VersionResps = new(VersionResp)
	if r, err := requests.New(&requests.Config{Timeout:time.Second * 30}); err == nil{
		hd :=  funcs.ServerHeader(config.VERSION, config.VERSIONCODE)
		if str, err := r.Get(fmt.Sprint(config.SERVERDOMAIN, "versions"), hd);err == nil{
			rsp := new(VersionResp)
			if err := json.Unmarshal([]byte(str), rsp); err == nil{
				VersionResps = rsp
			}
		}
	}
	return VersionResps
}

// 统一的格式化服务端返回的数据
func parseResp(dt []byte) (string, error){
	gr := gjson.ParseBytes(dt)
	mp := gr.Map()
	var msg string
	var code int
	if mp["msg"].Exists(){
		msg = mp["msg"].String()
	}
	if mp["code"].Exists(){
		code = int(mp["code"].Int())
	}
	if code == 200 && mp["data"].Exists(){
		return mp["data"].String(), nil
	}
	return "", errors.New(msg)
}

// 从服务端获取订阅代理
type SubResp struct{
	Id         int64    `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name       string   `json:"name" gorm:"default:null;index" form:"name"`         // 名称
	Remark     string   `json:"remark" gorm:"default:null;index" form:"remark"`     // 备注
	Local      string   `json:"local" gorm:"default:null;index"`                    // 地区
	Ip         string   `json:"ip" gorm:"default:null;index;"`                      // 代理的ip地址
	Timezone   string   `json:"timezone" gorm:"default:null;"`                      // 代理的时区
	Lang       string   `json:"lang" gorm:"default:null"`                           // 代理所在地区使用的语言
	Port       int      `json:"port" gorm:"index;default:0" form:"port"`            // 指定的端口,不存在则随机使用空余端口
	Config     string   `json:"config" gorm:"not null;" form:"config"`              // 代理信息,可以是vmess,vless等,也可以是http代理等
	ConfigMd5  string   `json:"config_md5" gorm:"uniqueIndex;not null"`             // 配置的md5,用于去重
	Username   string   `json:"username" gorm:"default:null;index" form:"username"` // 有些http代理等需要用户名
	Password   string   `json:"password" gorm:"default:null" form:"password"`       // 对应的密码
	Transfer   string   `json:"transfer" gorm:"default:null" form:"transfer"`       // 有些代理需要中转,无法直连.目的是解决有的好的ip在国外无法通过国内直连,可以是proxy的id或者具体配置
	AutoRun    int      `json:"auto_run" gorm:"default:0;index" form:"auto_run"`    // 系统启动跟随启动
	Private int `json:"private" gorm:"index;default:0;type:tinyint(1)"` // 是否是专属的
	Status int `json:"status" gorm:"index;type:tinyint(1);default:1"`// 状态
}
func GerProxySub(suburl string) ([]*SubResp, error){
	var rsp []*SubResp
	if r, err := requests.New(&requests.Config{Timeout:time.Second * 30}); err == nil{
		hd :=  funcs.ServerHeader(config.VERSION, config.VERSIONCODE)
		if bt, err := r.Get(suburl, hd);err == nil{
			respStr, err := parseResp(bt)
			if err != nil{
				return nil, err
			}
			if err := json.Unmarshal([]byte(respStr), &rsp); err != nil{
				return nil, err
			}
		}
	}
	return rsp, nil
}
