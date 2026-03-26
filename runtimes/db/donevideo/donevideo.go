package donevideo

type DoneVideo struct {
	ID       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID   int64  `json:"user_id" gorm:"index;default:0"`
	Filename string `json:"filename" gorm:"not null"`
	Cover    string `json:"cover"`
}
