package ws

import (
	"fmt"
	"tools/runtimes/eventbus"
)

func init() {
	// 版本信息通过api定时获取,不使用ws发送.原因是api的功能事先写好了,懒得改
	// eventbus.Bus.Subscribe("version", func(data any){

	// })

	// 意见或建议回复内容
	eventbus.Bus.Subscribe("sugge_cate", func(data any) {
		fmt.Println(data, "-----sugge_cate")
	})
}
