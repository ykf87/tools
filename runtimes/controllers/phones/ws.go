package phones

import (
	"net/http"
	"tools/runtimes/apptask"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/clients"
	"tools/runtimes/eventbus"
	"tools/runtimes/response"
	"tools/runtimes/ws"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
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
		if err := db.DB.Write(func(tx *gorm.DB) error {
			return phone.Save(phone, tx)
		}); err != nil {
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

	clients.TaskMgr.BindDevice(phone.DeviceId, &WSDelivery{DeviceId: phone.DeviceId})
	defer func() {
		clients.Hubs.Close(phone.DeviceId)           // 注销hub
		clients.TaskMgr.UnbindDevice(phone.DeviceId) // 注销任务绑定
	}()
	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		gs := gjson.ParseBytes(msg)
		tp := gs.Get("type").String()
		if tp != "" {
			if tp == "task_result" { // 报告任务执行情况
				go ReportTask(gs.Get("data").String(), phone.DeviceId)
				continue
			}
			dt, _ := config.Json.Marshal(map[string]any{
				"device_id": phone.DeviceId,
				"data":      gs.Get("data").String(),
			})
			go eventbus.Bus.Publish(gs.Get("type").String(), dt)
		}
	}
}

type TaskReport struct {
	RunId  int64  `json:"run_id"`
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

func ReportTask(str, deviceId string) {
	tr := new(TaskReport)
	if err := config.Json.Unmarshal([]byte(str), tr); err == nil {
		clients.TaskMgr.Report(tr.RunId, tr.Status, tr.Msg)
	}
}
