package tag

import (
	"tools/runtimes/db"

	"gorm.io/gorm"
)

type Tag struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name string `json:"name" gorm:"uniqueIndex" form:"name"`
}

func init() {
	db.DB.AutoMigrate(&Tag{})
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

// 通过id获取标签
func GetById(id any) *Tag {
	tg := new(Tag)
	db.DB.Model(&Tag{}).Where("id = ?", id).First(tg)
	return tg
}

// 通过标签名称获取对应的数组
func GetTagsByNames(names []string) map[string]int64 {
	var tgs []*Tag
	db.DB.Model(&Tag{}).Where("name in ?", names).Find(&tgs)
	mp := make(map[string]int64)
	for _, v := range tgs {
		mp[v.Name] = v.Id
	}
	return mp
}

// 通过id获取对应的数组
func GetTagsByIds(ids []int64) map[int64]string {
	var tgs []*Tag
	db.DB.Model(&Tag{}).Where("id in ?", ids).Find(&tgs)
	mp := make(map[int64]string)
	for _, v := range tgs {
		mp[v.Id] = v.Name
	}
	return mp
}
