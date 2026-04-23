package products

type Tag struct {
	ID    int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Langs []TagLang `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE"`
}

type TagLang struct {
	TagID int64  `json:"tag_id" gorm:"primaryKey;not null"`
	Lang  string `json:"lang" gorm:"primaryKey;size:10"`
}
