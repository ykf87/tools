package phones

import (
	"net/http"
	"tools/runtimes/db/clients"
	"tools/runtimes/response"
	"tools/runtimes/ws"

	"github.com/gin-gonic/gin"
)

func Ws(c *gin.Context) {
	deviceId := c.Query("device")
	phone, err := clients.GetPhoneByDeviceId(deviceId)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	if phone.Id < 1 {
		phone = &clients.Phone{
			DeviceId: deviceId,
			Brand:    c.Query("brand"),
			Version:  c.Query("version"),
			Os:       c.Query("os"),
		}
		if err := phone.Save(nil); err != nil { // 保存的时候需要判断是否允许自动加入,
			response.Error(c, http.StatusBadGateway, err.Error(), nil)
			return
		}
	}
	conn, err := ws.GetConn(c.Writer, c.Request, nil, func(r *http.Request) bool {
		return true
	})

	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	clients.Hubs.Register(phone.Id, phone.DeviceId, conn)
	// 将phone的分组设置到hubs
	for _, v := range phone.Tags {
		clients.Hubs.JoinGroup(v, phone.DeviceId)
	}
	for {
		_, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
