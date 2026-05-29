package products

import (
	"strconv"
	"strings"
	"tools/runtimes/storage"
)

func GetProductDetail(productID int64) (*SaveProductStruct, error) {

	var product Product
	if err := DB.DB().First(&product, productID).Error; err != nil {
		return nil, err
	}

	resp := &SaveProductStruct{
		ID:            product.ID,
		Spu:           product.Spu,
		Code:          product.Code,
		ProductType:   product.ProductType,
		Status:        product.Status,
		Width:         product.Width,
		Height:        product.Height,
		Length:        product.Length,
		BaseWeight:    product.BaseWeight,
		Weight:        product.Weight,
		PurchasePrice: product.PurchasePrice,
		SalePrice:     product.SalePrice,
		OriginPrice:   product.OriginPrice,
		Stock:         product.Stock,
	}

	// 1️⃣ Meta（多语言）
	meta, err := getProductMeta(productID)
	if err != nil {
		return nil, err
	}
	resp.Meta = meta

	// 2️⃣ 图片
	images, err := getProductImages(productID)
	if err != nil {
		return nil, err
	}
	resp.Images = images

	// 3️⃣ 普通属性
	attrs, skuAttrs, err := getProductAttrAndSkuAttrs(productID)
	if err != nil {
		return nil, err
	}
	resp.Attributes = attrs
	resp.SkuAttrs = skuAttrs

	// 4️⃣ SKU
	skus, err := getProductSKUs(productID)
	if err != nil {
		return nil, err
	}
	resp.Skus = skus
	resp.SkuAttrs = skuAttrs

	return resp, nil
}

func getProductMeta(productID int64) (map[string]SaveProductMeta, error) {

	var list []ProductInfo
	DB.DB().Where("product_id=?", productID).Find(&list)

	result := map[string]SaveProductMeta{}

	for _, m := range list {
		result[m.Lang] = SaveProductMeta{
			ProductID:      m.ProductID,
			Title:          m.Title,
			SubTitle:       m.SubTitle,
			Description:    m.Description,
			Content:        m.Content,
			SeoTitle:       m.SeoTitle,
			SeoDescription: m.SeoDescription,
			Keyword:        m.Keyword,
		}
	}

	return result, nil
}

func getProductImages(productID int64) ([]SaveProductImage, error) {

	var list []ProductImage
	DB.DB().Debug().Where("product_id=?", productID).Order("`index` ASC").Find(&list)

	var result []SaveProductImage

	for _, img := range list {
		result = append(result, SaveProductImage{
			Src: storage.Load("").URL(img.Src),
		})
	}

	return result, nil
}

func getProductAttrAndSkuAttrs(productID int64) ([]SaveProductAttrs, []SaveProductSkuAttr, error) {

	// 1️⃣ 查所有属性值（带 AttributeValue）
	var list []ProductAttributeValue
	err := DB.DB().
		Preload("Value").
		Where("product_id = ?", productID).
		Find(&list).Error
	if err != nil {
		return nil, nil, err
	}

	// 2️⃣ 查属性定义（判断 is_sku）
	var attrs []*ProductAttribute
	err = DB.DB().
		Where("product_id = ?", productID).
		Find(&attrs).Error
	if err != nil {
		return nil, nil, err
	}

	// for _, v := range attrs{

	// }

	// 👉 attr_id → is_sku
	attrTypeMap := map[int64]int8{}
	for _, a := range attrs {
		attrTypeMap[a.AttributeID] = a.IsSKU
	}

	// 3️⃣ 分组：attr_id → value list
	group := map[int64][]ProductAttributeValue{}

	for _, v := range list {
		group[v.AttributeID] = append(group[v.AttributeID], v)
	}

	// 4️⃣ 构建返回
	var normalAttrs []SaveProductAttrs
	var skuAttrs []SaveProductSkuAttr

	for attrID, values := range group {

		isSKU := attrTypeMap[attrID]

		// ---------- SKU属性 ----------
		if isSKU == 1 {

			var vlist []SaveProductAttrValue
			var isMedia bool

			for _, v := range values {

				valID := v.AttributeValueID

				item := SaveProductAttrValue{
					AttrValueID: valID,
				}

				// 👉 如果是媒体属性
				if v.IsMedia == 1 {

					if v.Images != "" {
						item.Images = strings.Split(v.Images, ",")
					}

					if v.Videos != "" {
						item.Videos = strings.Split(v.Videos, ",")
					}
					isMedia = true
				}

				vlist = append(vlist, item)
			}

			aid := attrID

			skuAttrs = append(skuAttrs, SaveProductSkuAttr{
				AttrID:  aid,
				Values:  vlist,
				Isimage: isMedia,
			})

		} else {
			// ---------- 普通属性 ----------

			var vids []int64
			for _, v := range values {
				vids = append(vids, v.AttributeValueID)
			}

			aid := attrID

			normalAttrs = append(normalAttrs, SaveProductAttrs{
				AttrID:      aid,
				AttrValueID: vids,
			})
		}
	}

	return normalAttrs, skuAttrs, nil
}

func getProductSKUs(productID int64) ([]SaveProductSku, error) {

	var skuList []ProductSKU
	if err := DB.DB().Where("product_id=?", productID).Find(&skuList).Error; err != nil {
		return nil, err
	}

	// SKU返回
	var skus []SaveProductSku

	// 👉 收集所有 value_id
	valueSet := map[int64]struct{}{}

	for _, s := range skuList {

		var ids []int64

		if s.AttrValueIDs != "" {
			parts := strings.Split(s.AttrValueIDs, ",")

			for _, p := range parts {
				id, err := strconv.ParseInt(p, 10, 64)
				if err != nil {
					continue
				}
				ids = append(ids, id)
				valueSet[id] = struct{}{}
			}
		}

		skus = append(skus, SaveProductSku{
			AttrValueIDs:  ids,
			SkuHash:       s.AttrHash,
			SalePrice:     s.Price,
			OriginPrice:   s.OriginPrice,
			PurchasePrice: s.CostPrice,
			Stock:         s.Stock,
			Code:          s.SkuCode,
			Status:        s.Status,
		})
	}
	return skus, nil

	// // 👉 没有SKU直接返回
	// if len(valueSet) == 0 {
	// 	return skus, nil
	// }

	// // 👉 查询 AttributeValue（关键）
	// var values []AttributeValue

	// var ids []int64
	// for id := range valueSet {
	// 	ids = append(ids, id)
	// }

	// if err := DB.DB().
	// 	Select("id, attribute_id").
	// 	Where("id IN ?", ids).
	// 	Find(&values).Error; err != nil {
	// 	return nil, nil, err
	// }

	// // 👉 构建 attr_id → value_id（去重）
	// attrMap := map[int64]map[int64]struct{}{}

	// for _, v := range values {

	// 	if _, ok := attrMap[v.AttributeID]; !ok {
	// 		attrMap[v.AttributeID] = map[int64]struct{}{}
	// 	}

	// 	attrMap[v.AttributeID][v.ID] = struct{}{}
	// }

	// // 👉 组装 SKU属性
	// var skuAttrs []SaveProductSkuAttr

	// for attrID, valMap := range attrMap {

	// 	var vlist []SaveProductAttrValue

	// 	for vid := range valMap {
	// 		vlist = append(vlist, SaveProductAttrValue{
	// 			AttrValueID: vid,
	// 		})
	// 	}

	// 	skuAttrs = append(skuAttrs, SaveProductSkuAttr{
	// 		AttrID: attrID,
	// 		Values: vlist,
	// 	})
	// }

	// return skus, skuAttrs, nil
}
