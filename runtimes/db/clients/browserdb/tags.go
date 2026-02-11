package browserdb

import (
	"tools/runtimes/db"

	"gorm.io/gorm"
)

type BrowserTag struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name string `json:"name" gorm:"index;not null" form:"name"`
	db.BaseModel
}

type BrowserToTag struct {
	BrowserId int64 `json:"browser_id" gorm:"primaryKey;" form:"browser_id"`
	TagId     int64 `json:"tag_id" gorm:"primaryKey" form:"tag_id"`
	db.BaseModel
}

// 删除标签
func (this *BrowserTag) Remove(tx *db.SQLiteWriter) error {
	if tx == nil {
		tx = db.DB
	}
	if this != nil && this.Id > 0 {
		err := tx.Write(func(txx *gorm.DB) error {
			return txx.Where("id = ?", this.Id).Delete(&BrowserTag{}).Error
		})
		if err != nil {
			return err
		}
		return tx.Write(func(txx *gorm.DB) error {
			return txx.Where("tag_id = ?", this.Id).Delete(&BrowserToTag{}).Error
		})
	}
	return nil
}

// 设置浏览器列表的tag
func SetBrowserTags(pcs []*Browser) {
	if len(pcs) < 1 {
		return
	}
	var ids []int64
	for _, v := range pcs {
		ids = append(ids, v.Id)
	}

	var pxtgs []*BrowserToTag
	db.DB.DB().Model(&BrowserToTag{}).Where("browser_id in ?", ids).Find(&pxtgs)

	var tagids []int64
	pcMap := make(map[int64][]int64)
	for _, v := range pxtgs {
		tagids = append(tagids, v.TagId)
		pcMap[v.BrowserId] = append(pcMap[v.BrowserId], v.TagId)
	}

	ttggs := GetBrowserTagsByIds(tagids)

	for _, v := range pcs {
		if ids, ok := pcMap[v.Id]; ok {
			for _, tid := range ids {
				if tgname, ok := ttggs[tid]; ok {
					v.Tags = append(v.Tags, tgname)
				}
			}
		}
	}
}

// 通过id获取对应的数组
func GetBrowserTagsByIds(ids []int64) map[int64]string {
	var tgs []*BrowserTag
	db.DB.DB().Model(&BrowserTag{}).Where("id in ?", ids).Find(&tgs)
	mp := make(map[int64]string)
	for _, v := range tgs {
		mp[v.Id] = v.Name
	}
	return mp
}

// 通过id获取标签
func GetBrowserTagsById(id any) *BrowserTag {
	tg := new(BrowserTag)
	db.DB.DB().Model(&BrowserTag{}).Where("id = ?", id).First(tg)
	return tg
}

// 通过标签名称获取对应的数组
func GetBrowserTagsByNames(names []string, tx *db.SQLiteWriter) map[string]int64 {
	if tx == nil {
		tx = db.DB
	}
	var tgs []*BrowserTag
	tx.DB().Model(&BrowserTag{}).Where("name in ?", names).Find(&tgs)
	mp := make(map[string]int64)
	for _, v := range tgs {
		mp[v.Name] = v.Id
	}

	var addn []*BrowserTag
	for _, v := range names {
		if _, ok := mp[v]; !ok {
			addn = append(addn, &BrowserTag{Name: v})
		}
	}
	if len(addn) > 0 {
		tx.Write(func(txx *gorm.DB) error {
			return txx.Create(&addn).Error
		})
	}
	for _, v := range addn {
		mp[v.Name] = v.Id
	}

	return mp
}

// 保存标签
// func (this *BrowserTag) Save(tx *db.SQLiteWriter) error {
// 	if tx == nil {
// 		tx = db.DB
// 	}
// 	if this.Id > 0 {
// 		return tx.Write(func(txx *gorm.DB) error {
// 			return txx.Model(&BrowserTag{}).Where("id = ?", this.Id).
// 				Updates(map[string]any{
// 					"name": this.Name,
// 				}).Error
// 		})
// 	} else {
// 		return tx.Write(func(txx *gorm.DB) error {
// 			return txx.Create(this).Error
// 		})
// 	}
// }

// 删除某个Browser下的tag
func (this *Browser) RemoveMyTags(tx *db.SQLiteWriter) error {
	if tx == nil {
		tx = db.DB
	}
	return tx.Write(func(txx *gorm.DB) error {
		return txx.Where("browser_id = ?", this.Id).Delete(&BrowserToTag{}).Error
	})
}

// 使用当前的tag标签完全替换已有标签
// 使用此方法会清空已有的tag
func (this *Browser) CoverTgs(tagsName []string, tx *db.SQLiteWriter) error {
	if tx == nil {
		tx = db.DB
	}

	if err := this.RemoveMyTags(tx); err != nil {
		return err
	}

	mp := GetBrowserTagsByNames(tagsName, tx)

	var ntag []BrowserToTag
	for _, tagg := range tagsName {
		if tid, ok := mp[tagg]; ok {
			ntag = append(ntag, BrowserToTag{BrowserId: this.Id, TagId: tid})
		}
	}

	if len(ntag) > 0 {
		if err := tx.Write(func(txx *gorm.DB) error {
			return txx.Create(ntag).Error
		}); err != nil {
			return err
		}
	}

	return nil
}
