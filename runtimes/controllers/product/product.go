package product

import (
	"fmt"
	"net/http"
	"strconv"
	"tools/runtimes/db"
	"tools/runtimes/db/products"
	"tools/runtimes/response"

	"github.com/gin-gonic/gin"
)

func GetList(c *gin.Context) {
	var req db.ListFinder
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	ps, total, err := products.GetProductList(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response.Success(c, gin.H{
		"list":  ps,
		"total": total,
	}, "")
}

// 获取属性
func GetProductAttrs(c *gin.Context) {
	var q db.ListFinder

	if err := c.ShouldBindJSON(&q); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	attrs, err := products.GetAttributeList(q)
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	var data any
	if q.Lang != "" {
		data = products.BuildAttributeDTO(attrs, q.Lang)
	} else {
		data = products.BuildAttributeMap(attrs)
	}

	response.Success(c, data, "")
}

func SaveProduct(c *gin.Context) {
	var req products.SaveProductStruct
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	if err := products.SaveProduct(req); err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	response.Success(c, nil, "")
}

func GetProductRow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		fmt.Println("-----")
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}

	pi, err := products.GetProductDetail(int64(id))
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error(), nil)
		return
	}
	response.Success(c, pi, "")
}
