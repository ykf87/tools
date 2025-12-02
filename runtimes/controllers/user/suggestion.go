package user

import "github.com/gin-gonic/gin"

type SuggestionData struct {
	Content string `json:"content" form:"content"`
}

// 接收反馈信息
func Suggestion(c *gin.Context) {

}
