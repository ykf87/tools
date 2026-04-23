package products

type ProductSKUAttribute struct {
	ID          int64 `gorm:"primaryKey"`
	ProductID   int64 `gorm:"uniqueIndex:idx_product_sku_attr"`
	AttributeID int64 `gorm:"uniqueIndex:idx_product_sku_attr"` // 复用你前面的 Attribute
	Sort        int   `gorm:"default:0"`
	// 可选值（红、蓝、L、XL）
	Values []ProductSKUAttrValue `gorm:"foreignKey:SKUAttrID;constraint:OnDelete:CASCADE"`
}

type ProductSKUAttrValue struct {
	ID        int64 `gorm:"primaryKey"`
	SKUAttrID int64 `gorm:"index"`
	// 复用属性值（推荐）
	AttributeValueID int64 `gorm:"index"`
	// 可扩展：自定义值（比如刻字选项）
	// CustomValue string `gorm:"size:255"`
	Sort int `gorm:"default:0"`
}

type ProductSKU struct {
	ID             int64  `gorm:"primaryKey"`
	ProductID      int64  `gorm:"index"`
	SKUCode        string `gorm:"size:100;index"` // 商家编码
	Price          int64  `gorm:"not null"`       // 分
	Stock          int64  `gorm:"default:0"`
	CombinationKey string `gorm:"size:255;uniqueIndex"`
	SoldCount      int64  `gorm:"default:0"`
	// SKU组合
	Options []ProductSKUOption `gorm:"foreignKey:SKUID;constraint:OnDelete:CASCADE"`
	// 可定制数据（JSON，核心）
	// CustomData datatypes.JSON `gorm:"type:jsonb"`
	Status int8 `gorm:"default:1;index"`
}

type ProductSKUOption struct {
	ID               int64 `gorm:"primaryKey"`
	SKUID            int64 `gorm:"uniqueIndex:idx_sku_attr"`
	AttributeID      int64 `gorm:"uniqueIndex:idx_sku_attr"`
	AttributeValueID int64 `gorm:"index"`
	// 可选：自定义（比如刻字）
	// CustomValue string `gorm:"size:255"`
}
