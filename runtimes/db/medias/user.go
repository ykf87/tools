package medias

import (
	"fmt"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"

	"gorm.io/gorm"
)

type MediaUser struct {
	Id       int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string          `json:"name" gorm:"index;default:null"`     // 用户名
	Cover    string          `json:"cover" gorm:"default:null"`          // 头像
	Platform string          `json:"platform" gorm:"index:plu;not null"` // 怕太
	Uuid     string          `json:"uuid" gorm:"index:plu;not null"`     // 访问主页等
	Account  string          `json:"account" gorm:"index;"`              // 例如抖音号,用于用户搜索的
	AdminID  int64           `json:"admin_id" gorm:"index;default:0"`    // 哪个后台用户添加的
	Addtime  int64           `json:"addtime" gorm:"default:0;index"`     // 添加时间
	Works    int64           `json:"works" gorm:"index;default:-1"`      // 发布作品数量
	Fans     int64           `json:"fans" gorm:"index;default:-1"`       // 粉丝数
	Local    string          `json:"local" gorm:"index;default:null"`    // 所在地区
	Tags     []string        `json:"tags" gorm:"-"`                      // 标签
	Clients  map[int][]int64 `json:"clients" gorm:"-"`                   // 使用的客户端
	Proxys   []int64         `json:"proxys" gorm:"-"`                    // 使用的代理列表
}

type MediaUserToTag struct {
	UserID int64 `json:"user_id" gorm:"primaryKey;not null"`
	TagID  int64 `json:"tag_id" gorm:"primaryKey;not null"`
}

// 媒体用户自动获取时使用的浏览器
type MediaUserToClient struct {
	MUID       int64 `json:"media_user" gorm:"primaryKey;not null"`
	ClientType int   `json:"client_type" gorm:"primaryKey;not null"`
	ClientID   int64 `json:"client_id" gorm:"primaryKey;not null"`
}

// 媒体用户自动获取时使用的代理,设置多个代理将随机选取代理
type MediaUserProxy struct {
	MUID    int64 `json:"mu_id" gorm:"primaryKey;not null"`
	ProxyID int64 `json:"proxy_id" gorm:"primaryKey;not null"`
}

func (this *MediaUser) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}

	if strings.HasPrefix(this.Cover, config.MediaUrl) {
		if af, ok := strings.CutPrefix(this.Cover, config.MediaUrl); ok {
			this.Cover = af
		}
	}
	if this.Id > 0 {
		err := tx.Model(&MediaUser{}).Where("id = ?", this.Id).
			Updates(map[string]any{
				"platform": this.Platform,
				"name":     this.Name,
				"cover":    this.Cover,
				"uuid":     this.Uuid,
				"fans":     this.Fans,
				"works":    this.Works,
				"local":    this.Local,
				"account":  this.Account,
			}).Error
		if err != nil {
			return err
		}
		// eventbus.Bus.Publish("media_save", this)
		return nil
	} else {
		this.Addtime = time.Now().Unix()
		return tx.Create(this).Error
	}
}

func (this *MediaUser) GetTags() []*MediaUserTag {
	var tagIDs []int64
	dbs.Model(&MediaUserToTag{}).Select("tag_id").Where("user_id = ?", this.Id).Find(&tagIDs)

	var tags []*MediaUserTag
	dbs.Model(&MediaUserTag{}).Where("id in ?", tagIDs).Find(&tags)
	return tags
}

func (this *MediaUser) GetClients() map[int][]int64 {
	var rows []*MediaUserToClient
	dbs.Model(&MediaUserToClient{}).Where("m_uid = ?", this.Id).Find(&rows)
	this.Clients = make(map[int][]int64)
	for _, v := range rows {
		this.Clients[v.ClientType] = append(this.Clients[v.ClientType], v.ClientID)
	}
	return this.Clients
}

// 补全媒体用户的tag和客户端
func (this *MediaUser) Commpare() {
	this.GetClients()
	this.GetProxys()
	this.GenAvatarToHttp()

	this.Tags = nil
	for _, zv := range this.GetTags() {
		this.Tags = append(this.Tags, zv.Name)
	}
}

func (this *MediaUser) GenAvatarToHttp() {
	if !strings.HasPrefix(this.Cover, "http") {
		this.Cover = fmt.Sprintf("%s/%s", config.MediaUrl, this.Cover)
	}
}

// 获取代理列表
func (this *MediaUser) GetProxys() []int64 {
	var ids []int64
	dbs.Model(&MediaUserProxy{}).Select("proxy_id").Where("m_uid = ?", this.Id).Find(&ids)
	this.Proxys = ids
	return this.Proxys
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
		v.Commpare()
		// if v.Cover != "" {
		// 	if !strings.HasPrefix(v.Cover, "http") {
		// 		v.Cover = fmt.Sprintf("%s/%s", config.MediaUrl, v.Cover)
		// 	}
		// }
	}
	return mus, total
}

// 根据id获取用户信息
func GetMediaUserByID(id any) *MediaUser {
	mu := new(MediaUser)
	if err := dbs.Model(&MediaUser{}).Where("id = ?", id).First(mu).Error; err != nil || mu.Id < 1 {
		return nil
	}
	// mu.Commpare()
	return mu
}

func (this *MediaUser) EmptyClient(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}
	return tx.Where("m_uid = ?", this.Id).Delete(&MediaUserToClient{}).Error
}
func (this *MediaUser) EmptyProxy(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}
	return tx.Where("m_uid = ?", this.Id).Delete(&MediaUserProxy{}).Error
}
func (this *MediaUser) EmptyTag(tx *gorm.DB) error {
	if tx == nil {
		tx = dbs
	}
	return tx.Where("user_id = ?", this.Id).Delete(&MediaUserToTag{}).Debug().Error
}
