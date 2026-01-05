package task

import (
	"net/http"
	"strconv"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/clients"
	"tools/runtimes/db/tasks"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type TaskData struct {
	Page  int    `json:"page" form:"page"`
	Limit int    `json:"limit" form:"limit"`
	Q     string `json:"q" form:"q"`
}

func List(c *gin.Context) {
	var user *admins.Admin
	if u, ok := c.Get("_user"); ok {
		if user, ok = u.(*admins.Admin); !ok {
			response.Error(c, http.StatusBadRequest, "请登录", nil)
			return
		}
	}
	if user == nil || user.Id < 1 {
		response.Error(c, http.StatusBadRequest, "请登录", nil)
		return
	}

	dt := new(TaskData)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	list, total := tasks.GetTasks(dt.Page, dt.Limit, dt.Q, user.Id)
	rsp := gin.H{
		"list":  list,
		"total": total,
	}

	response.Success(c, rsp, "")
}

// 后台任务列表基础数据
func BaseData(c *gin.Context) {
	rsp := map[string]any{
		"tags": tasks.GetTags(),
		"tps":  tasks.Types,
		"devices": map[string]any{
			"web":   clients.GetAllBrowsers(1, 20, ""),
			"phone": clients.GetAllPhones(1, 20, ""),
		},
	}

	rrs, _ := parses.Marshal(rsp, c)
	response.Success(c, rrs, "")
}

// 添加或编辑任务
func AddOrEdit(c *gin.Context) {
	var user *admins.Admin
	if u, ok := c.Get("_user"); ok {
		if user, ok = u.(*admins.Admin); !ok {
			response.Error(c, http.StatusBadRequest, "请登录", nil)
			return
		}
	}
	if user == nil || user.Id < 1 {
		response.Error(c, http.StatusBadRequest, "请登录", nil)
		return
	}

	dt := new(tasks.Task)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if dt.Starttime > 0 {
		dt.Starttime = dt.Starttime / 1000
	}
	if dt.Endtime > 0 {
		dt.Endtime = dt.Endtime / 1000
	}
	dt.AdminId = user.Id
	newDevices := dt.Devices

	if err := dt.Save(nil); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if dt.ID > 0 {
		dt.RemoveNotUsedDevices(dt.Devices)
	}
	var dvs []*tasks.TaskClients
	for _, v := range newDevices {
		dvs = append(dvs, &tasks.TaskClients{
			TaskID:     dt.ID,
			DeviceType: dt.Tp,
			DeviceID:   v,
		})
	}
	if len(dvs) > 0 {
		db.TaskDB.Create(dvs)
	}
	response.Success(c, dt, "")
}

// 获取任务下的设备列表
func TaskDevices(c *gin.Context) {
	taskId, _ := strconv.Atoi(c.Query("id"))
	if taskId < 1 {
		response.Error(c, http.StatusBadRequest, "错误", nil)
		return
	}

	task := tasks.GetTaskById(taskId)
	if task == nil || task.ID < 1 {
		response.Error(c, http.StatusBadRequest, "错误!", nil)
		return
	}

	dids := task.GetDevices()

	var dvs any
	switch task.Tp {
	case 0:
		dvs = clients.GetBrowsersByIds(dids)
	case 1:
		dvs = clients.GetPhonesByIds(dids)
	}
	response.Success(c, dvs, "")
}
