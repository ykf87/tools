package products

import "gorm.io/datatypes"

// 定制配置
type ProductCustomAttribute struct {
	ID        int64 `json:"id" gorm:"primaryKey"`
	ProductID int64 `json:"product_id" gorm:"index"`

	// 可复用属性表
	AttributeID int64 `gorm:"index"`

	// 类型（非常关键）
	Type int8 `gorm:"index"`
	// 1=选择型（复用AttributeValue）
	// 2=文本
	// 3=图片
	// 4=3D贴图
	Values []ProductCustomAttrValue `gorm:"foreignKey:CustomAttrID"`
	Rule   ProductCustomConfig      `gorm:"foreignKey:CustomAttrID"`

	Required bool `gorm:"default:false"` // 是否必填
	Sort     int  `gorm:"default:0"`
}

// 选择型定制（复用 AttributeValue）
type ProductCustomAttrValue struct {
	ID           int64 `gorm:"primaryKey"`
	CustomAttrID int64 `gorm:"index"`

	// 复用属性值
	AttributeValueID int64 `gorm:"index"`
	ResourceID       int64 `gorm:"index"`

	Sort int `gorm:"default:0"`
}

// 定制规则表（重点）
type ProductCustomConfig struct {
	ID           int64 `gorm:"primaryKey"`
	CustomAttrID int64 `gorm:"index"`
	// 文本限制
	MaxLength int
	MinLength int

	// 图片限制
	MaxImages int
	MaxSizeKB int

	// 3D/贴图限制
	MaxCount int

	// 可选：JSON扩展
	Config datatypes.JSON `gorm:"type:jsonb"`
}
