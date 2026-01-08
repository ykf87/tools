package js

import (
	"net/http"
	"tools/runtimes/db"
	"tools/runtimes/db/admins"
	"tools/runtimes/db/jses"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	dt := new(db.ListFinder)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	lst, total := jses.GetJsList(dt)
	response.Success(c, gin.H{
		"list":  lst,
		"total": total,
	}, "")
}

func Tags(c *gin.Context) {
	response.Success(c, jses.GetTags(), "")
}

func GetTypes(c *gin.Context) {
	response.Success(c, jses.Types, "")
}

func Delete(c *gin.Context) {
	err := jses.Delete(c.Query("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
	} else {
		response.Success(c, nil, "")
	}
}

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

	dt := new(jses.Js)
	if err := c.ShouldBind(dt); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	dt.AdminID = user.Id
	if err := dt.Save(nil); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// js参数
	if err := dt.GenParams(); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 任务标签处理
	if len(dt.Tags) > 0 {
		dt.AddTags()
	}
	response.Success(c, dt, "")
}
