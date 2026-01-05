package jses

import "tools/runtimes/db"

type JsTag struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"type:varchar(60);index"`
}

type JsToTag struct {
	JsID  int64 `json:"js_id" gorm:"primaryKey;not null"`
	TagID int64 `json:"tag_id" gorm:"primaryKey;not null"`
}

func init() {
	db.DB.AutoMigrate(&JsTag{})
	db.DB.AutoMigrate(&JsToTag{})
}
