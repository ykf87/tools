// app端触发的一些订阅事件
package clients

import "tools/runtimes/eventbus"

func init() {
	eventbus.Bus.Subscribe("init", func(data any) { // 收到的设备验证消息,用于判断是否能保持连接

	})
}
