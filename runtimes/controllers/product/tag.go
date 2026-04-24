package product

import (
	"net/http"
	"tools/runtimes/db/products"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

type tagreq struct {
	Page  int    `json:"page" form:"page"`
	Limit int    `json:"limit" form:"limit"`
	Lang  string `json:"lang" form:"lang"`
	Q     string `json:"q" form:"q"`
}
type TagDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func Tag(c *gin.Context) {
	var req tagreq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	var list []TagDTO

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	products.DB.DB().Table("tags t").
		Select(`
			t.id,
			COALESCE(tl1.name, tl2.name) as name
		`).
		Joins("LEFT JOIN tag_langs tl1 ON tl1.tag_id = t.id AND tl1.lang = ?", req.Lang).
		Joins("LEFT JOIN tag_langs tl2 ON tl2.tag_id = t.id AND tl2.lang = ?", "zh-CN").
		Scan(&list)
	response.Success(c, list, "")
}
