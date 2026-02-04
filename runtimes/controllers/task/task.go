package task

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/clients"
	"tools/runtimes/db/clients/browserdb"
	"tools/runtimes/db/jses"
	"tools/runtimes/db/tasks"
	"tools/runtimes/parses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

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

	dt := new(db.ListFinder)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	list, total := tasks.GetTasks(user.Id, dt)
	rsp := gin.H{
		"list":  list,
		"total": total,
	}

	response.Success(c, rsp, "")
}

// 后台任务列表基础数据
func BaseData(c *gin.Context) {
	finder := &db.ListFinder{
		Page:  1,
		Limit: 1000,
		Q:     "",
	}
	jjs, _ := jses.GetJsList(finder)
	for _, v := range jjs {
		v.Name = fmt.Sprintf("%s:%s", v.Name, strings.Join(v.Tags, ","))

		switch v.Tp {
		case 0:
			v.Name = "[浏览器]" + v.Name
		case 1:
			v.Name = "[手机端]" + v.Name
		case 2:
			v.Name = "[HTTP]" + v.Name
		}
	}

	rsp := map[string]any{
		"tags": tasks.GetTags(),
		"tps":  tasks.Types,
		"devices": map[string]any{
			"web":   browserdb.GetAllBrowsers(1, 20, ""),
			"phone": clients.GetAllPhones(1, 20, ""),
		},
		"jss": jjs,
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

	dt.AdminId = user.Id
	if err := dt.Save(nil); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 任务设备处理
	dt.GenDevices()

	// 参数处理
	dt.GenParams(dt.Params)

	// 任务标签处理
	if len(dt.Tags) > 0 {
		dt.AddTags()
	}
	// fmt.Println(dt.Params, "---")
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
		dvs = browserdb.GetBrowsersByIds(dids)
	case 1:
		dvs = clients.GetPhonesByIds(dids)
	}
	response.Success(c, dvs, "")
}

func Delete(c *gin.Context) {
	if err := tasks.DeleteByID(c.Query("id")); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	// 停止任务 todo...
	response.Success(c, nil, "删除成功")
}

func RuningTasks(c *gin.Context) {
	// if admin, err := admins.GetAdminUser(c); err == nil {
	// 	response.Success(c, tasks.GetRuningTasks(admin.Id), "")
	// 	return
	// }
	response.Error(c, http.StatusBadRequest, "", nil)
}
