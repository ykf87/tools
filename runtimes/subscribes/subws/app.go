package subws

import (
	"fmt"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db/clients"
	"tools/runtimes/eventbus"

	"github.com/tidwall/gjson"
)

type DeviceScreen struct {
	Width   int `json:"width"`
	Height  int `json:"height"`
	Density int `json:"density"`
}
type DeviceInfo struct {
	Brand   string       `json:"brand"`
	Model   string       `json:"model"`
	System  string       `json:"system"`
	Version string       `json:"release"`
	Screen  DeviceScreen `json:"screen"`
}

func init() {
	connFirst()
	listenPhoneStatus()
}

func connFirst() {
	eventbus.Bus.Subscribe("init", func(data any) { // 收到的设备验证消息,用于判断是否能保持连接
		if dtbt, ok := data.([]byte); ok {
			gs := gjson.ParseBytes(dtbt)
			deviceId := gs.Get("device_id").String()

			phone, err := clients.GetPhoneByDeviceId(deviceId)
			if err != nil {
				closeWs(deviceId, err.Error())
				return
			}
			if phone.Status != 1 {
				msg := "设置状态异常"
				if phone.CloseMsg != "" {
					msg = phone.CloseMsg
				}
				closeWs(deviceId, msg)
				return
			}

			info := new(DeviceInfo)
			if err := config.Json.Unmarshal([]byte(gs.Get("data").String()), info); err != nil {
				closeWs(deviceId, err.Error())
				return
			}

			ifbt, _ := config.Json.Marshal(info)
			phone.Brand = info.Brand + ": " + info.Model
			phone.Os = info.System
			phone.Version = info.Version
			phone.Infos = string(ifbt)
			go phone.Save(nil)

			// 发送设备编号
			num := map[string]any{
				"type": "devicenum",
				"data": phone.Num,
			}
			bt, _ := config.Json.Marshal(num)
			fmt.Println("发送设备编号----")
			clients.Hubs.SentClient(phone.DeviceId, bt)
		}
	})
}

func listenPhoneStatus() {
	eventbus.Bus.Subscribe("phone-status-change", func(data any) {
		if phone, ok := data.(*clients.Phone); ok {
			if phone.Status != 1 {
				go closeWs(phone.DeviceId, phone.CloseMsg)
			}
		}
	})
}

func closeWs(deviceid string, msg string) {
	mp := map[string]any{
		"type": "close",
		"data": map[string]any{
			"reconnect": false,
			"msg":       msg,
		},
	}
	fmt.Println("错误消息-------准备关闭ws")
	if dt, err := config.Json.Marshal(mp); err == nil {
		fmt.Println("发送关闭ws")
		clients.Hubs.SentClient(deviceid, dt)
	}
	time.Sleep(time.Microsecond * 100)
	clients.Hubs.Close(deviceid)
}
