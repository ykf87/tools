package notifies

type Notify struct {
	Id int64 `json:"id" gorm:"primaryKey;autoIncrement"`
}
