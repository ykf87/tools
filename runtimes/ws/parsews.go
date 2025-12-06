package ws

import (
	"fmt"
	"tools/runtimes/config"
	"tools/runtimes/eventbus"
)

type Uuid struct {
	Id         int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	Uuid       string  `json:"uuid" gorm:"uniqueIndex;not null"` // 唯一id
	Cost       float64 `json:"cost" gorm:"index;default:0.0"`    // 总共支付金额
	Addtime    int64   `json:"addtime" gorm:"index;default:0"`
	Updatetime int64   `json:"updatetime" gorm:"index;default:0"`
	Inviter    int64   `json:"inviter" gorm:"index;default:0"`                // 邀请人id,也就是user表的id,表明这个设备是通过这个user邀请来的
	RealIp     string  `json:"real_ip" gorm:"default:0;index"`                // 真实的客户端ip地址
	Status     int     `json:"status" gorm:"default:1;index;type:tinyint(1)"` // 1为正常状态
	StatusMsg  string  `json:"status_msg" gorm:"default:null"`                // 账号status不是1的原因
}

func init() {
	// 版本信息通过api定时获取,不使用ws发送.原因是api的功能事先写好了,懒得改
	// eventbus.Bus.Subscribe("version", func(data any){

	// })

	// 意见或建议回复内容
	eventbus.Bus.Subscribe("sugge_cate", func(data any) {
		fmt.Println(data, "-----sugge_cate")
	})

	// 服务端uuid状态改变收到的消息
	eventbus.Bus.Subscribe("uuid_status_change", func(data any) {
		if dtstr, ok := data.(string); ok {
			uid := new(Uuid)
			if err := config.Json.Unmarshal([]byte(dtstr), uid); err == nil {
				if uid.Id > 0 {
					config.MainStatus = uid.Status
					config.MainStatusMsg = uid.StatusMsg
				}
			}
		}
	})
}
