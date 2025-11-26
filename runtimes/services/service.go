package services

import (
	"encoding/json"
	"fmt"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
	"tools/runtimes/requests"
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

// func init(){// 这段代码在web.go中执行
// 	for{
// 		GetVersions()
// 		time.Sleep(time.Hour)
// 	}
// }

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
