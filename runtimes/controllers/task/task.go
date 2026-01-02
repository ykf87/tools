package task

import (
	"net/http"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/tasks"
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
	}
	response.Success(c, rsp, "")
}

// 添加或编辑任务
func AddOrEdit(c *gin.Context) {
	dt := new(tasks.Task)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := dt.Save(nil); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, dt, "")
}
