package medias

import (
	"fmt"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/eventbus"

	"gorm.io/gorm"
)

type MediaUser struct {
	Id         int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string   `json:"name" gorm:"index;default:null"`                   // 用户名
	Cover      string   `json:"cover" gorm:"default:null"`                        // 头像
	Platform   string   `json:"platform" gorm:"index:plu;not null"`               // 怕太
	Uuid       string   `json:"uuid" gorm:"index:plu;not null"`                   // 访问主页等
	Account    string   `json:"account" gorm:"index;"`                            //例如抖音号,用于用户搜索的
	AdminID    int64    `json:"admin_id" gorm:"index;default:0"`                  // 哪个后台用户添加的
	Addtime    int64    `json:"addtime" gorm:"default:0;index"`                   // 添加时间
	Works      int64    `json:"works" gorm:"index;default:-1"`                    // 发布作品数量
	Fans       int64    `json:"fans" gorm:"index;default:-1"`                     // 粉丝数
	Local      string   `json:"local" gorm:"index;default:null"`                  // 所在地区
	Tags       []string `json:"tags" gorm:"-"`                                    // 标签
	ClientType int      `json:"client_type" gorm:"int;type:tinyint(1);default:0"` // 客户端类型,0-浏览器 1-手机
	ClientID   int64    `json:"client_id" gorm:"index;default:0"`                 // 客户端
}

type MediaUserToTag struct {
	UserID int64 `json:"user_id" gorm:"primaryKey;not null"`
	TagID  int64 `json:"tag_id" gorm:"primaryKey;not null"`
}

func (this *MediaUser) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}

	if this.Id > 0 {
		err := tx.Model(&Media{}).Where("id = ?", this.Id).
			Updates(map[string]any{
				"platform": this.Platform,
				"name":     this.Name,
				"cover":    this.Cover,
				"uuid":     this.Uuid,
			}).Error
		if err != nil {
			return err
		}
		eventbus.Bus.Publish("media_save", this)
		return nil
	} else {
		this.Addtime = time.Now().Unix()
		return tx.Create(this).Error
	}
}

func GetUserPlatforms() map[string]string {
	var pls []string
	dbs.Model(&MediaUser{}).Select("platform").Group("platform").Find(&pls)

	plsmap := make(map[string]string)
	for _, v := range pls {
		plsmap[v] = v
	}
	return plsmap
}

func (this *MediaUser) GetTags() []*MediaUserTag {
	var tagIDs []int64
	dbs.Model(&MediaUserToTag{}).Select("tag_id").Where("user_id = ?", this.Id).Find(&tagIDs)

	var tags []*MediaUserTag
	dbs.Model(&MediaUserTag{}).Where("id in ?", tagIDs).Find(&tags)
	return tags
}

func GetMediaUsers(adminID int64, dt *db.ListFinder) ([]*MediaUser, int64) {
	var mus []*MediaUser
	if dt.Page < 1 {
		dt.Page = 1
	}
	if dt.Limit < 1 {
		dt.Limit = 20
	}
	md := dbs.Model(&MediaUser{}).Where("admin_id = ?", adminID)
	if dt.Q != "" {
		qs := fmt.Sprintf("%%%s%%", dt.Q)
		md.Where("name like ?", qs)
	}

	if len(dt.Tags) > 0 {
		var muids []int64
		dbs.Model(&MediaUserToTag{}).Select("user_id").Where("tag_id in ?", dt.Tags).Find(&muids)
		if len(muids) > 0 {
			md.Where("id in ?", muids)
		}
	}

	var total int64
	md.Count(&total)

	if dt.Scol != "" && dt.By != "" {
		var byy string
		if strings.Contains(dt.By, "desc") {
			byy = "desc"
		} else {
			byy = "asc"
		}
		md.Order(fmt.Sprintf("%s %s", dt.Scol, byy))
	}
	md.Order("id DESC").Offset((dt.Page - 1) * dt.Limit).Limit(dt.Limit).Find(&mus)

	for _, v := range mus {
		// v.Devices = v.GetDevices()
		if v.Cover != "" {
			v.Cover = fmt.Sprintf("%s/%s", config.MediaUrl, v.Cover)
		}
		for _, zv := range v.GetTags() {
			v.Tags = append(v.Tags, zv.Name)
		}
		// v.GetParams()
	}
	return mus, total
}
