// 启动的必须检查
package runtimes

import (
	"fmt"
	"time"
	"tools/runtimes/config"
	suggestions "tools/runtimes/db/Suggestions"
	"tools/runtimes/db/clients"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"
	"tools/runtimes/requests"

	"github.com/tidwall/gjson"
)

func init() {
	config.RuningRoot = funcs.RunnerPath()

	GetVersions()
}

// 如果有什么参数需要在开始初始化的时候获取,在此次添加
func GetVersions() {
	if r, err := requests.New(&requests.Config{Timeout: time.Second * 10}); err == nil {
		hd := funcs.ServerHeader(config.VERSION, config.VERSIONCODE)
		if respbt, err := r.Get(fmt.Sprint(config.SERVERDOMAIN, "start"), hd); err == nil {
			gs := gjson.ParseBytes(respbt)
			if gs.Get("code").Int() != 200 {
				config.MainStatus = 0
				config.MainStatusMsg = gs.Get("msg").String()
			} else {
				gsd := gs.Get("data")
				suggestions.UpSuggCateFromServer(gsd.Get("sugg_cate").String())
				if err := config.VersionsFromServer(gsd.Get("versions").String()); err != nil {
					logs.Error(err.Error())
				}
				if gsd.Get("phone_max").Int() > 0 {
					clients.MaxPhoneNum = gsd.Get("phone_max").Int()
				}
			}
		}
	}
}
