package product

import (
	"net/http"
	"tools/runtimes/db/products"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func UpsertAttributes(c *gin.Context) {
	var req []products.MultiLangAttr
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := products.UpsertAttributes(req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.Success(c, nil, "")
}

// 删除属性,算了，属性不能删
type rmvids struct {
	IDs []int64 `json:"ids" form:"ids"`
}

// func RemoveAttr(c *gin.Context) {
// 	var req rmvids
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response.Error(c, http.StatusBadRequest, err.Error(), nil)
// 		return
// 	}

// 	products.DB.DB().Model(&products.AttributeValue)
// }
