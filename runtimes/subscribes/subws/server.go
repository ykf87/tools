package subws

// 接收服务端发送来的ws消息的处理

import (
	"fmt"
	"tools/runtimes/config"
	suggestions "tools/runtimes/db/Suggestions"
	"tools/runtimes/db/messages"
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
	cguuid()
	sugg()
	serverInfor()
	serverNotify()
}

// uuid 状态改变事件
func cguuid() {
	eventbus.Bus.Subscribe("server_uuid_status_change", func(data any) {
		if dtstr, ok := data.(string); ok {
			uid := new(Uuid)

			if err := config.Json.Unmarshal([]byte(dtstr), uid); err == nil {
				if uid.Id > 0 {
					config.MainStatus = uid.Status // 此状态如果不为1,在web请求的中间件中执行拦截
					config.MainStatusMsg = uid.StatusMsg
					msg := new(messages.Message)
					msg.Duration = 10000
					if uid.Status == 1 {
						msg.Content = "账号已恢复"
						msg.Type = "success"
					} else {
						msg.Type = "error"
						if uid.StatusMsg == "" {
							uid.StatusMsg = "账号状态异常"
						}
						msg.Content = uid.StatusMsg
					}
					msg.Send()
				}
			}
		}
	})
}

// 意见或建议分类
func sugg() {
	eventbus.Bus.Subscribe("server_sugge_cate", func(data any) {
		if dtstr, ok := data.(string); ok {
			var suc *suggestions.SuggCate
			if err := config.Json.Unmarshal([]byte(dtstr), suc); err == nil {
				suc.Save(nil)
			}
		}
	})
}

// 接收服务端消息通知
func serverInfor() {
	eventbus.Bus.Subscribe("server_information", func(data any) {
		fmt.Println(data, "-----sugge_cate")
	})
}

// 接收服务端发来的通知 notify
func serverNotify() {
	eventbus.Bus.Subscribe("server_notify", func(data any) {
		if dtstr, ok := data.(string); ok {
			var suc *suggestions.SuggCate
			if err := config.Json.Unmarshal([]byte(dtstr), suc); err == nil {
				suc.Save(nil)
			}
		}
	})
}
