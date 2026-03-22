package users

import (
	"errors"
	"net/http"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db/medias"
	"tools/runtimes/db/messages"
	"tools/runtimes/listens/ws"
	"tools/runtimes/response"
	"tools/runtimes/runner"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// 从平台获取用户的主页视频列表
func OnlinVideo(c *gin.Context) {
	type tmp struct {
		UserID int64  `json:"user_id"`
		Lang   string `json:"lang"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Page   int    `json:"page"`
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

	// url, js, err := mu.GetJsAndUrl("info")
	// if err != nil {
	// 	response.Error(c, http.StatusBadRequest, err.Error(), nil)
	// 	return
	// }

	// opts := runner.GenWebOpt(
	// 	mainsignal.MainCtx,
	// 	-1,
	// 	true,
	// 	url,
	// 	js,
	// 	nil,
	// 	0,
	// 	req.Width,
	// 	req.Height,
	// 	req.Lang,
	// 	"",
	// )
	// fmt.Println("js:", js)
	_, bid := mu.GetCanUseClient()
	opts, err := mu.GenBrowserOpt(bid, req.Page)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if opts.Width == 0 {
		opts.Width = req.Width
	}
	if opts.Height == 0 {
		opts.Height = req.Height
	}
	if opts.Language == "" {
		opts.Language = req.Lang
	}

	// opts.Show = false

	r, err := runner.GetRunner(0, opts)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	// response.Success(c, nil, "内容获取中...")
	// return
	go func() {
		type UserVids struct {
			Url   string `json:"url"`
			Cover string `json:"cover"`
			Title string `json:"title"`
			Like  int64  `json:"like"`
			Href  string `json:"href"`
		}
		err = r.Start(time.Second*60, func(msg, s string) error {
			gs := gjson.Parse(s)
			if gs.Get("lists").Exists() {
				var lists []*UserVids
				if err := config.Json.Unmarshal([]byte(gs.Get("lists").String()), &lists); err == nil {
					works := gs.Get("works").Int()
					if bt, err := config.Json.Marshal(map[string]any{
						"type": "onlinevideo",
						"data": map[string]any{
							"total": works,
							"list":  lists,
						},
					}); err == nil {
						ws.Broadcost(bt)
					}
				}

				// for _, v := range gs.Get("lists").Array() {
				// 	lists = append(lists, v.String())
				// }

				return nil
			}
			return errors.New("错误的内容")
		}, func(msg string) {

		}, func(msg string) {

		})
		if err != nil {
			messages.ErrorMsg(err.Error())
		} else {
			messages.SuccessMsg("获取成功")
		}
	}()
	response.Success(c, nil, "内容获取中...")
}
