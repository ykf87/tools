package products

import (
	"tools/runtimes/db"
)

// 商品基础表
type Product struct {
	ID            int64                    `json:"id" gorm:"primaryKey;autoIncrement"`
	Spu           string                   `json:"spu" gorm:"uniqueIndex;not null;"`                                   // 产品型号(唯一)
	Code          string                   `json:"code" gorm:"index;default:null"`                                     // 商家编码
	Meta          []ProductInfo            `json:"meta" gorm:"foreignKey:ProductID"`                                   // 产品信息
	Images        []ProductImage           `json:"images" gorm:"default:null;foreignKey:ProductID"`                    // 图集
	Videos        []ProductVideo           `json:"videos" gorm:"default:null;foreignKey:ProductID"`                    // 视频集
	Weight        int64                    `json:"weight" gorm:"default:0"`                                            // 重量,单位克
	Width         int64                    `json:"width" gorm:"default:0"`                                             // 宽(cm)
	Height        int64                    `json:"height" gorm:"default:0"`                                            // 高(cm)
	Length        int64                    `json:"length" gorm:"default:0"`                                            // 长
	Brand         int64                    `json:"brand" gorm:"index;default:0"`                                       // 品牌
	PublishAt     int64                    `json:"publish_at" gorm:"default:0;index"`                                  // 定时上架
	ProductType   int                      `json:"product_type" gorm:"type:tinyint(1);default:0"`                      // 商品类型,0实物, 1定制，2虚拟
	OriginPrice   int64                    `json:"origin_price" gorm:"default:0"`                                      // 原价,单位分
	SalePrice     int64                    `json:"sale_price" gorm:"default:0"`                                        // 售价,单位分
	PurchasePrice int64                    `json:"purchase_price" gorm:"default:0"`                                    // 进货价,成本价,单位分
	Stock         int64                    `json:"stock" gorm:"default:0;index"`                                       // 库存,0为不限
	Customer      []ProductCustomAttribute `json:"customer" gorm:"foreignKey:ProductID"`                               // 定制配置
	Tags          []Tag                    `json:"tags" gorm:"many2many:product_tag_relations;"`                       // 标签列表
	Attributes    []ProductAttribute       `json:"attributes" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"` // 产品属性
	Skus          []ProductSKU             `json:"skus" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`       // SKU列表
	SKUAttrs      []ProductSKUAttribute    `json:"sku_attrs" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`  // SKU维度（销售属性）
	Status        int8                     `json:"status" gorm:"default:1;index"`                                      // 1启用 0禁用
}

type ProductImage struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID int64  `json:"product_id" gorm:"index;not null;"`
	Src       string `json:"src" gorm:"not null;"`
	Index     int    `json:"index" gorm:"index;default:0"` // 排序
}

type ProductVideo struct {
	ID        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID int64  `json:"product_id" gorm:"index;not null;"`
	Src       string `json:"src" gorm:"not null"`
	Cover     string `json:"cover" gorm:"default:null"`    // 视频封面
	Index     int    `json:"index" gorm:"index;default:0"` // 排序
}

// 商品详情表
type ProductInfo struct {
	ProductID      int64  `json:"-" gorm:"primaryKey;not null;uniqueIndex:idx_product_id_lang"` // 产品id
	Lang           string `json:"lang" gorm:"primaryKey;uniqueIndex:idx_product_id_lang"`       // 语言
	Title          string `json:"title" gorm:"not null;index"`                                  // 产品标题
	SubTitle       string `json:"sub_title" gorm:"index;default:null"`                          // 短标题
	Description    string `json:"description" gorm:"default:null"`                              // 产品简洁
	Content        string `json:"content" gorm:"default:null"`                                  // 产品详情
	SeoTitle       string `json:"seo_title" gorm:"default:null;index"`                          // seo标题
	SeoDescription string `json:"seo_description" gorm:"default:null;index"`                    // seo简介
	Keyword        string `json:"keyword" gorm:"index;default:null"`                            // 关键词
}

var DB = db.PRODUCTDB

const (
	DEFLANG = "zh-CN"
)

func init() {
	DB.DB().AutoMigrate(
		&Product{},
		&ProductImage{},
		&ProductVideo{},
		&ProductInfo{},
		&Tag{},
		&TagLang{},
		&Attribute{},
		&AttributeLang{},
		&AttributeValue{},
		&AttributeValueLang{},
		&ProductAttribute{},
		&ProductAttributeValue{},
		&ProductAttrIndex{},
		&ProductSKUAttribute{},
		&ProductSKUAttrValue{},
		&ProductSKU{},
		&ProductSKUOption{},
		&ProductCustomAttribute{},
		&ProductCustomAttrValue{},
		&ProductCustomConfig{},
	)
}

// 获取产品列表
func GetProductList(req db.ListFinder) ([]Product, int64, error) {
	var ps []Product
	var total int64

	query := DB.DB().Model(&Product{}).
		Where("status = ?", 1)

	// 关键词
	if req.Q != "" {
		query = query.Joins("JOIN product_infos pi ON pi.product_id = products.id").
			Where("pi.title ILIKE ?", "%"+req.Q+"%")
	}

	// 标签筛选
	if len(req.Tags) > 0 {
		query = query.Joins("JOIN product_tag_relations ptr ON ptr.product_id = products.id").
			Where("ptr.tag_id IN ?", req.Tags)
	}

	// // 属性筛选（重点）
	// if req.Filters != nil {
	// 	query = query.Joins("JOIN product_attributes pa ON pa.product_id = products.id").
	// 		Joins("JOIN product_attribute_values pav ON pav.product_attribute_id = pa.id")

	// 	var conditions []string
	// 	var args []interface{}

	// 	for k, f := range req.Filters {
	// 		conditions = append(conditions, "(pa.attribute_id = ? AND pav.attribute_value_id IN ?)")
	// 		args = append(args, f.AttributeID, f.ValueIDs)
	// 	}

	// 	query = query.Where(strings.Join(conditions, " OR "), args...).
	// 		Group("products.id").
	// 		Having("COUNT(DISTINCT pa.attribute_id) = ?", len(q.Filters))
	// }

	// // 价格筛选（SKU）
	// if q.MinPrice > 0 || q.MaxPrice > 0 {
	// 	query = query.Joins("JOIN product_skus ps ON ps.product_id = products.id")

	// 	if q.MinPrice > 0 {
	// 		query = query.Where("ps.price >= ?", q.MinPrice)
	// 	}
	// 	if q.MaxPrice > 0 {
	// 		query = query.Where("ps.price <= ?", q.MaxPrice)
	// 	}
	// }

	// 统计总数
	query.Count(&total)

	// 分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	offset := (req.Page - 1) * req.Limit

	err := query.
		Preload("Images").
		Preload("Videos").
		Preload("Meta").
		Preload("Skus").
		Order("products.id DESC").
		Limit(req.Limit).
		Offset(offset).
		Find(&ps).Error

	return ps, total, err
}
