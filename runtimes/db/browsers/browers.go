package browsers

type Browser struct {
	Id int64 `json:"id" gorm:"primaryKey;autoIncrement"`
}
