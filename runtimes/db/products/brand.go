package products
type Brand struct{
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	Logo string `json:"logo" gorm:"default:null"` // 品牌图标
}
