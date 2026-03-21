package users

import (
	"fmt"
	"net/http"
	"tools/runtimes/db/medias"
	"tools/runtimes/mainsignal"
	"tools/runtimes/response"
	"tools/runtimes/runner"

	"github.com/gin-gonic/gin"
)

// 从平台获取用户的主页视频列表
func OnlinVideo(c *gin.Context) {
	type tmp struct {
		UserID int64  `json:"user_id"`
		Lang   string `json:"lang"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}

	var req tmp
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	mu := medias.GetMediaUserByID(req.UserID)
	if mu == nil || mu.Id < 1 {
		response.Error(c, http.StatusBadRequest, "用户不存在", nil)
		return
	}

	url, js, err := mu.GetJsAndUrl()
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	opts := runner.GenWebOpt(
		mainsignal.MainCtx,
		-1,
		true,
		url,
		js,
		nil,
		0,
		req.Width,
		req.Height,
		req.Lang,
		"",
	)
	// fmt.Println("js:", js)

	r, err := runner.GetRunner(0, opts)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	r.Start(0, func(s string) error {
		fmt.Println(s)
		return nil
	})
}
