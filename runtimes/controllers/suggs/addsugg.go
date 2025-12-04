package suggs

import (
	"net/http"
	"tools/runtimes/db"
	suggestions "tools/runtimes/db/Suggestions"
	"tools/runtimes/response"
	"tools/runtimes/services"

	"github.com/gin-gonic/gin"
)

type AddSuggStruct struct {
	CateId  int64  `json:"cate"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func AddSuggestion(c *gin.Context) {
	adddata := new(AddSuggStruct)
	if err := c.ShouldBindJSON(adddata); err != nil {
		response.Error(c, http.StatusNotFound, err.Error(), nil)
		return
	}
	if adddata.CateId < 1 {
		response.Error(c, http.StatusBadRequest, "请选择 反馈类别", nil)
		return
	}
	if adddata.Title == "" {
		response.Error(c, http.StatusBadRequest, "请填写 简要描述", nil)
		return
	}

	sug := new(suggestions.Suggestion)
	sug.CateId = adddata.CateId
	sug.Title = adddata.Title
	sug.Content = adddata.Content

	tx := db.DB.Begin()
	if err := sug.Save(tx); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	// 发送给服务端
	if err := services.AddSuggestion(sug); err != nil {
		tx.Rollback()
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	tx.Commit()
	response.Success(c, nil, "Suggestion submited.")
}
