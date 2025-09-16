package db

type Sys struct {
	Id int64 `json:"id" gorm:"primaryKey;autoIncrement"`
}
