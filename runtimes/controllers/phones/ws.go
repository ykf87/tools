package phones

import (
	"net/http"
	"tools/runtimes/apptask"
	"tools/runtimes/config"
	"tools/runtimes/db/clients"
	"tools/runtimes/eventbus"
	"tools/runtimes/response"
	"tools/runtimes/ws"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type WSDelivery struct {
	DeviceId string
}

func (this *WSDelivery) Mode() string {
	return "ws"
}
func (this *WSDelivery) Deliver(task *apptask.AppTask) error {
	bt, err := config.Json.Marshal(task)
	if err != nil {
		return err
	}
	clients.Hubs.SentClient(this.DeviceId, bt)
	return nil
}
func (this *WSDelivery) Pick(deviceId string) *apptask.AppTask {
	return nil
}

func Ws(c *gin.Context) {
	deviceId := c.Query("device")
	phone, _ := clients.GetPhoneByDeviceId(deviceId)
	// if err != nil {
	// 	response.Error(c, http.StatusBadGateway, err.Error(), nil)
	// 	return
	// }
	if phone == nil || phone.Id < 1 {
		phone = &clients.Phone{
			DeviceId: deviceId,
			Brand:    c.Query("brand"),
			Version:  c.Query("version"),
			Os:       c.Query("os"),
		}
		if err := phone.Save(nil); err != nil {
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
	defer clients.Hubs.Close(phone.DeviceId)
	// clients.Hubs.SentClient()
	// clients.TaskMgr.BindDevice(phone.DeviceId, apptask.WithWS(conn))

	clients.TaskMgr.BindDevice(phone.DeviceId, &WSDelivery{DeviceId: phone.DeviceId})
	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		gs := gjson.ParseBytes(msg)
		if gs.Get("type").String() != "" {
			dt, _ := config.Json.Marshal(map[string]any{
				"device_id": phone.DeviceId,
				"data":      gs.Get("data").String(),
			})
			go eventbus.Bus.Publish(gs.Get("type").String(), dt)
		}
	}
}
