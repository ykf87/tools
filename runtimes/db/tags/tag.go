package tags

import (
	"tools/runtimes/db"

	"gorm.io/gorm"
)

type Tag struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"uniqueIndex"`
}

func (this *Tag) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Model(&Tag{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"name": this.Name,
			}).Error
	} else {
		return tx.Create(this).Error
	}
}

// 删除
func (this *Tag) Remove(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this != nil && this.Id > 0 {
		return db.DB.Where("id = ?", this.Id).Delete(&Tag{}).Error
	}
	return nil
}

func GetById(id any) *Tag {
	tg := new(Tag)
	db.DB.Model(&Tag{}).Where("id = ?", id).First(tg)
	return tg
}
