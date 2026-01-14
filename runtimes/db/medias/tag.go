package medias

type MediaUserTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"index;not null"`
}

func GetTags() []*MediaUserTag {
	var tags []*MediaUserTag
	dbs.Model(&MediaUserTag{}).Find(&tags)
	return tags
}
