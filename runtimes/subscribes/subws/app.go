package subws

import "tools/runtimes/eventbus"

func init() {
	connFirst()
}

func connFirst() {
	eventbus.Bus.Subscribe("init", func(data any) { // 收到的设备验证消息,用于判断是否能保持连接

	})
}
