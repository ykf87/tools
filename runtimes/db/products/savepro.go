package products

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"tools/runtimes/storage"

	"gorm.io/gorm"
)

// 保存产品相关
type SaveProductMeta struct {
	ProductID      int64  `json:"product_id" form:"product_id"`
	Title          string `json:"title" form:"title"`
	SubTitle       string `json:"sub_title" form:"sub_title"`
	Description    string `json:"description" form:"description"`
	Content        string `json:"content" form:"content"`
	SeoTitle       string `json:"seo_title" form:"seo_title"`
	SeoDescription string `json:"seo_description" form:"seo_description"`
	Keyword        string `json:"keyword" form:"keyword"`
}
type SaveProductImage struct {
	Src   string `json:"src" form:"src"`
	Index int    `json:"index" form:"index"`
}
type SaveProductVideo struct {
	ID    int64  `json:"id" form:"id"`
	Src   string `json:"src" form:"src"`
	Cover string `json:"cover" form:"cover"`
	Index int    `json:"index" form:"index"`
}
type SaveProductAttrValue struct {
	AttrValueID int64    `json:"attr_value_id" form:"attr_value_id"`
	Images      []string `json:"images" form:"images"`
	Videos      []string `json:"videos" form:"videos"`
}
type SaveProductSkuAttr struct {
	AttrID  int64                  `json:"attr_id" form:"attr_id"`
	Isimage bool                   `json:"isimage" form:"isimage"`
	Values  []SaveProductAttrValue `json:"values" form:"values"`
}
type SaveProductSku struct {
	AttrValueIDs  []int64  `json:"attr_value_ids" form:"attr_value_ids"`
	SkuHash       string   `json:"sku_hash" form:"sku_hash"`
	SalePrice     float64  `json:"sale_price" form:"sale_price"`
	OriginPrice   float64  `json:"origin_price" form:"origin_price"`
	PurchasePrice float64  `json:"purchase_price" form:"purchase_price"`
	Stock         int64    `json:"stock" form:"stock"`
	Images        []string `json:"images" form:"images"`
	Videos        []string `json:"videos" form:"videos"`
	Code          string   `json:"code" form:"code"`
	Status        int      `json:"status" form:"status"`
}
type SaveProductAttrs struct {
	AttrID      int64   `json:"attr_id" form:"attr_id"`
	AttrValueID []int64 `json:"attr_value_id" form:"attr_value_id"`
}
type SaveProductStruct struct {
	ID            int64                      `json:"id" form:"id" binding:"omitempty"`
	Spu           string                     `json:"spu" form:"spu"`
	Code          string                     `json:"code" form:"code"`
	Weight        int64                      `json:"weight" form:"weight"`
	BaseWeight    int64                      `json:"base_weight" form:"base_weight"`
	Width         int64                      `json:"width" form:"width"`
	Height        int64                      `json:"height" form:"height"`
	Length        int64                      `json:"length" form:"length"`
	Brand         int64                      `json:"brand" form:"brand"`
	ProductType   int                        `json:"product_type" form:"product_type"`
	OriginPrice   float64                    `json:"origin_price" form:"origin_price"`
	SalePrice     float64                    `json:"sale_price" form:"sale_price"`
	PurchasePrice float64                    `json:"purchase_price" form:"purchase_price"`
	Stock         int64                      `json:"stock" form:"stock"`
	Status        int                        `json:"status" form:"status"`
	Meta          map[string]SaveProductMeta `json:"meta" form:"meta"`
	Images        []SaveProductImage         `json:"images" form:"images"`
	Tags          []int64                    `json:"tags" form:"tags"`
	SkuAttrs      []SaveProductSkuAttr       `json:"sku_attrs" form:"sku_attrs"`
	Skus          []SaveProductSku           `json:"skus" form:"skus"`
	Attributes    []SaveProductAttrs         `json:"attributes" form:"attributes"`
}

func SaveProduct(req SaveProductStruct) error {
	return DB.Write(func(tx *gorm.DB) error {

		// 1️⃣ 保存 Product
		productID, err := saveProductBase(tx, req)
		if err != nil {
			fmt.Println(err, "======1")
			return err
		}

		// 2️⃣ 多语言
		if err := saveProductMeta(tx, productID, req.Meta); err != nil {
			fmt.Println(err, "======2")
			return err
		}

		// 3️⃣ 媒体
		if err := saveProductImages(tx, productID, req.Images); err != nil {
			fmt.Println(err, "======3")
			return err
		}

		// 4️⃣ 属性绑定
		if err := saveProductAttributes(tx, productID, req.Attributes, req.SkuAttrs); err != nil {
			fmt.Println(err, "======4")
			return err
		}

		// 5️⃣ 普通属性值
		if err := saveProductAttrValues(tx, productID, req.Attributes, req.SkuAttrs); err != nil {
			fmt.Println(err, "======5")
			return err
		}

		// 6️⃣ SKU（🔥核心）
		if err := saveSKUs(tx, productID, req.Skus); err != nil {
			fmt.Println(err, "======6")
			return err
		}

		// 7️⃣ 同步冗余
		if err := SyncProductData(tx, productID); err != nil {
			fmt.Println(err, "======7")
			return err
		}

		// 8️⃣ 同步筛选
		if err := SyncProductFilterIndex(tx, productID); err != nil {
			fmt.Println(err, "======8")
			return err
		}

		return nil
	})
}

func saveProductBase(tx *gorm.DB, req SaveProductStruct) (int64, error) {
	p := Product{
		ID:            req.ID,
		Spu:           req.Spu,
		Code:          req.Code,
		ProductType:   req.ProductType,
		Status:        req.Status,
		PurchasePrice: req.PurchasePrice,
		Weight:        req.Weight,
		Height:        req.Height,
		Length:        req.Length,
		Width:         req.Width,
		OriginPrice:   req.OriginPrice,
		SalePrice:     req.SalePrice,
		Stock:         req.Stock,
	}

	if p.ID > 0 {
		return p.ID, tx.Model(&p).Updates(p).Error
	} else {
		var product Product
		tx.Where("spu = ?", req.Spu).First(&product)
		if product.ID > 0 {
			p.ID = product.ID
			return p.ID, tx.Model(&p).Updates(p).Error
		}
	}

	if err := tx.Create(&p).Error; err != nil {
		return 0, err
	}

	return p.ID, nil
}

func saveSKUs(tx *gorm.DB, productID int64, skus []SaveProductSku) error {
	fmt.Println(skus, "---")
	// 查已有 SKU
	var oldSKUs []ProductSKU
	tx.Where("product_id=?", productID).Find(&oldSKUs)

	oldMap := map[string]ProductSKU{}
	for _, s := range oldSKUs {
		oldMap[s.AttrHash] = s
	}

	newMap := map[string]SaveProductSku{}
	for _, s := range skus {
		newMap[s.SkuHash] = s
	}

	// 1️⃣ 删除旧的（不存在的）
	for hash, old := range oldMap {
		if _, ok := newMap[hash]; !ok {
			tx.Delete(&old)
		}
	}

	// 2️⃣ 新增 / 更新
	for _, s := range skus {

		attrIDs := make([]string, len(s.AttrValueIDs))
		for i, v := range s.AttrValueIDs {
			attrIDs[i] = fmt.Sprintf("%d", v)
		}

		sort.Strings(attrIDs)

		hash := strings.Join(attrIDs, "_")

		sku := ProductSKU{
			ProductID:    productID,
			AttrHash:     hash,
			AttrValueIDs: strings.Join(attrIDs, ","),
			Price:        s.SalePrice,
			OriginPrice:  s.OriginPrice,
			CostPrice:    s.PurchasePrice,
			Stock:        int64(s.Stock),
			SkuCode:      s.Code,
			Status:       s.Status,
		}

		var existing ProductSKU
		err := tx.Where("product_id=? AND attr_hash=?", productID, hash).
			First(&existing).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 新增
			if err := tx.Create(&sku).Error; err != nil {
				return err
			}

			// 写组合表
			for _, vid := range s.AttrValueIDs {
				tx.Create(&SKUAttributeValue{
					SKUID:            sku.ID,
					AttributeValueID: int64(vid),
				})
			}

		} else {
			// 更新
			tx.Model(&existing).Updates(sku)
		}
	}

	return nil
}

func saveProductAttributes(tx *gorm.DB, productID int64, attrs []SaveProductAttrs, skuAttrs []SaveProductSkuAttr) error {

	// 清空重建（这里可以接受）
	tx.Where("product_id=?", productID).Delete(&ProductAttribute{})

	// 普通属性
	for _, a := range attrs {
		tx.Create(&ProductAttribute{
			ProductID:   productID,
			AttributeID: int64(a.AttrID),
			IsSKU:       0,
		})
	}

	// SKU属性
	for _, s := range skuAttrs {
		tx.Create(&ProductAttribute{
			ProductID:   productID,
			AttributeID: int64(s.AttrID),
			IsSKU:       1,
		})
	}

	return nil
}

func saveProductImages(tx *gorm.DB, productID int64, imgs []SaveProductImage) error {
	tx.Where("product_id=?", productID).Delete(&ProductImage{})

	for idx, img := range imgs {
		tx.Create(&ProductImage{
			ProductID: productID,
			Src:       storage.Load("").Base(img.Src),
			Index:     idx,
		})
	}

	return nil
}

func saveProductMeta(tx *gorm.DB, productID int64, meta map[string]SaveProductMeta) error {

	// 删除旧的（简单稳定）
	if err := tx.Where("product_id = ?", productID).
		Delete(&ProductInfo{}).Error; err != nil {
		return err
	}

	var list []ProductInfo

	for lang, m := range meta {
		list = append(list, ProductInfo{
			ProductID: productID,
			Lang:      lang,

			Title:       m.Title,
			SubTitle:    m.SubTitle,
			Description: m.Description,
			Content:     m.Content,

			SeoTitle:       m.SeoTitle,
			SeoDescription: m.SeoDescription,
			Keyword:        m.Keyword,
		})
	}

	if len(list) == 0 {
		return nil
	}

	return tx.Create(&list).Error
}

func saveProductAttrValues(tx *gorm.DB, productID int64, attrs []SaveProductAttrs, skuAttrs []SaveProductSkuAttr) error {

	// 删除旧数据
	if err := tx.Where("product_id = ?", productID).
		Delete(&ProductAttributeValue{}).Error; err != nil {
		return err
	}

	var list []ProductAttributeValue

	// 1️⃣ 普通属性（无图）
	for _, a := range attrs {

		if a.AttrID < 1 {
			continue
		}

		for _, vid := range a.AttrValueID {

			list = append(list, ProductAttributeValue{
				ProductID:        productID,
				AttributeID:      a.AttrID,
				AttributeValueID: vid,
				IsMedia:          0,
			})
		}
	}

	// 2️⃣ SKU属性（处理图片，比如颜色）
	for _, sa := range skuAttrs {

		if sa.AttrID < 1 {
			continue
		}

		for _, v := range sa.Values {

			if v.AttrValueID < 1 {
				continue
			}

			// 👉 只在有图时才写
			// if len(v.Images) == 0 && len(v.Videos) == 0 {
			// 	continue
			// }

			pav := ProductAttributeValue{
				ProductID:        productID,
				AttributeID:      sa.AttrID,
				AttributeValueID: v.AttrValueID,
			}
			if sa.Isimage {
				pav.IsMedia = 1
				if len(v.Images) > 0 {
					pav.Images = strings.Join(v.Images, ",")
				}
				if len(v.Videos) > 0 {
					pav.Videos = strings.Join(v.Videos, ",")
				}
			}

			list = append(list, pav)
		}
	}

	if len(list) == 0 {
		return nil
	}

	return tx.Debug().Create(&list).Error
}
