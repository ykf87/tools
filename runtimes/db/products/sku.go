package products

import (
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// type ProductSKUAttribute struct {
// 	ID          int64 `gorm:"primaryKey"`
// 	ProductID   int64 `gorm:"uniqueIndex:idx_product_sku_attr"`
// 	AttributeID int64 `gorm:"uniqueIndex:idx_product_sku_attr"` // 复用你前面的 Attribute
// 	Sort        int   `gorm:"default:0"`
// 	// 可选值（红、蓝、L、XL）
// 	Values []ProductSKUAttrValue `gorm:"foreignKey:SKUAttrID;constraint:OnDelete:CASCADE"`
// }

// type ProductSKUAttrValue struct {
// 	ID        int64 `gorm:"primaryKey"`
// 	SKUAttrID int64 `gorm:"index"`
// 	// 复用属性值（推荐）
// 	AttributeValueID int64 `gorm:"index"`
// 	// 可扩展：自定义值（比如刻字选项）
// 	Sort int `gorm:"default:0"`
// }

// type ProductSKU struct {
// 	ID             int64  `gorm:"primaryKey"`
// 	ProductID      int64  `gorm:"index"`
// 	SKUCode        string `gorm:"size:100;index"` // 商家编码
// 	Price          int64  `gorm:"not null"`       // 分
// 	Stock          int64  `gorm:"default:0"`
// 	CombinationKey string `gorm:"size:255;uniqueIndex"`
// 	SoldCount      int64  `gorm:"default:0"`
// 	// SKU组合
// 	Options []ProductSKUOption `gorm:"foreignKey:SKUID;constraint:OnDelete:CASCADE"`
// 	// 可定制数据（JSON，核心）
// 	// CustomData datatypes.JSON `gorm:"type:jsonb"`
// 	Status int8 `gorm:"default:1;index"`
// }

// type ProductSKUOption struct {
// 	ID               int64 `gorm:"primaryKey"`
// 	SKUID            int64 `gorm:"uniqueIndex:idx_sku_attr"`
// 	AttributeID      int64 `gorm:"uniqueIndex:idx_sku_attr"`
// 	AttributeValueID int64 `gorm:"index"`
// 	// 可选：自定义（比如刻字）
// 	// CustomValue string `gorm:"size:255"`
// }

type ProductSKU struct {
	ID        int64 `json:"id" gorm:"primaryKey"`
	ProductID int64 `json:"product_id" gorm:"not null;index"`

	SkuCode string `json:"sku_code" gorm:"size:100;uniqueIndex"`

	// 🔥 SKU唯一组合
	AttrHash string `json:"attr_hash" gorm:"size:255;not null"`

	// 🔥 冗余（加速查询）
	AttrValueIDs string `json:"attr_value_ids" gorm:"size:255;index"`

	// 💰 价格
	Price       int64 `json:"price" gorm:"not null;index"`
	OriginPrice int64 `json:"origin_price"`
	CostPrice   int64 `json:"cost_price"`

	// 📦 库存
	Stock int64 `json:"stock" gorm:"index"`
	Sales int64 `json:"sales"`

	// 📸 SKU图片
	Image string `json:"image" gorm:"size:255"`

	Weight int64 `json:"weight"`

	Status int8 `json:"status" gorm:"default:1;index"`

	// 🔥 组合唯一（防重复SKU）
	_ struct{} `gorm:"uniqueIndex:uk_product_attrhash,priority:1"`
}

type SKUAttributeValue struct {
	ID int64 `json:"id" gorm:"primaryKey"`

	SKUID int64 `json:"sku_id" gorm:"not null;index:idx_sku_attr,priority:1"`

	AttributeID      int64 `json:"attribute_id" gorm:"not null;index"`
	AttributeValueID int64 `json:"attribute_value_id" gorm:"not null;index:idx_sku_attr,priority:2"`

	// 👉 防重复（一个SKU不能有两个相同属性）
	_ struct{} `gorm:"uniqueIndex:uk_sku_attr,priority:1"`
}

// SKU 自动生成（核心算法）
// 👉 输入：属性值二维数组（颜色×尺寸）
func CartesianProduct(arr [][]int64) [][]int64 {
	if len(arr) == 0 {
		return nil
	}

	result := [][]int64{{}}

	for _, values := range arr {
		var temp [][]int64
		for _, r := range result {
			for _, v := range values {
				newCombo := append([]int64{}, r...)
				newCombo = append(newCombo, v)
				temp = append(temp, newCombo)
			}
		}
		result = temp
	}

	return result
}

// 生成 SKU
func GenerateSKUs(db *gorm.DB, productID int64, attrValues [][]int64) error {
	combos := CartesianProduct(attrValues)

	for _, combo := range combos {
		sort.Slice(combo, func(i, j int) bool {
			return combo[i] < combo[j]
		})

		// 生成 hash
		hash := make([]string, len(combo))
		for i, v := range combo {
			hash[i] = fmt.Sprintf("%d", v)
		}
		attrHash := strings.Join(hash, "_")
		attrIDs := strings.Join(hash, ",")

		sku := ProductSKU{
			ProductID:    productID,
			AttrHash:     attrHash,
			AttrValueIDs: attrIDs,
			Price:        0,
			Stock:        0,
		}

		// 🔥 防重复（依赖唯一索引）
		err := db.Where("product_id=? AND attr_hash=?", productID, attrHash).
			FirstOrCreate(&sku).Error
		if err != nil {
			return err
		}

		// 写 SKUAttributeValue
		for _, valID := range combo {
			db.FirstOrCreate(&SKUAttributeValue{
				SKUID:            sku.ID,
				AttributeValueID: valID,
			})
		}
	}

	return nil
}
