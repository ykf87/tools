package clients

import (
	"time"
	"tools/runtimes/browser"
	"tools/runtimes/db"
	"tools/runtimes/db/proxys"
	"tools/runtimes/eventbus"
	"tools/runtimes/proxy"

	"gorm.io/gorm"
)

type Browser struct {
	Id          int64  `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name        string `json:"name" gorm:"index;not null" form:"name"`                // 名称
	Proxy       int64  `json:"proxy" gorm:"index;default:0" form:"proxy"`             // 代理
	ProxyName   string `json:"proxy_name" gorm:"index;default:null" form:"-"`         // 代理的名称,只有设置了代理id才有
	ProxyConfig string `json:"proxy_config" gorm:"default:null;" form:"proxy_config"` // 代理的配置项,如果设置了此值,proxy失效
	Local       string `json:"local" gorm:"default:null" form:"local"`                // 所在地区
	Lang        string `json:"lang" gorm:"default:null;" form:"lang"`                 // 语言
	Timezone    string `json:"timezone" gorm:"default:null;" form:"timezone"`         // 时区
	Width       int    `json:"width" gorm:"default:1920" form:"width"`                // 屏幕宽度
	Height      int    `json:"height" gorm:"default:1040" form:"height"`              // 屏幕高度
	Ip          string `json:"ip" gorm:"default:null;"`                               // ip地址,设置了代理才有
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        []string      `json:"tags" gorm:"-" form:"tags"` // 标签
	Bs          *browser.User `json:"-" gorm:"-" form:"-"`       // 浏览器
	Opend       bool          `json:"opend" gorm:"-" form:"-"`   // 是否启动
}

type BrowserTag struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name string `json:"name" gorm:"index;not null" form:"name"`
}

type BrowserToTag struct {
	BrowserId int64 `json:"browser_id" gorm:"primaryKey;" form:"browser_id"`
	TagId     int64 `json:"tag_id" gorm:"primaryKey" form:"tag_id"`
}

func init() {
	db.DB.AutoMigrate(&Browser{})
	db.DB.AutoMigrate(&BrowserTag{})
	db.DB.AutoMigrate(&BrowserToTag{})

	// 监听代理改变事件,同步修改浏览器的local
	eventbus.Bus.Subscribe("proxy_change", func(dt any) {
		if proxy, ok := dt.(*proxys.Proxy); ok {
			go func() {
				time.Sleep(time.Second * 1)
				db.DB.Model(&Browser{}).Where("proxy = ?", proxy.Id).Updates(map[string]any{
					"local":      proxy.Local,
					"ip":         proxy.Ip,
					"lang":       proxy.Lang,
					"timezone":   proxy.Timezone,
					"proxy_name": proxy.Name,
				})
			}()
		}
	})

	eventbus.Bus.Subscribe("browser-close", func(dt any) {
		if bu, ok := dt.(*browser.User); ok {
			if bu.Id > 0 {
				if bs, err := GetBrowserById(bu.Id); err == nil {
					eventbus.Bus.Publish("ws", map[string]any{
						"browser": bs,
					})
				}
			}
		}
	})
}

// 标签操作开始-----------------------------------
// 保存标签
func (this *BrowserTag) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Model(&BrowserTag{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"name": this.Name,
			}).Error
	} else {
		return tx.Create(this).Error
	}
}

// 删除标签
func (this *BrowserTag) Remove(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this != nil && this.Id > 0 {
		err := tx.Where("id = ?", this.Id).Delete(&BrowserTag{}).Error
		if err != nil {
			return err
		}
		return tx.Where("tag_id = ?", this.Id).Delete(&BrowserToTag{}).Error
	}
	return nil
}

// 通过id获取标签
func GetBrowserTagsById(id any) *BrowserTag {
	tg := new(BrowserTag)
	db.DB.Model(&BrowserTag{}).Where("id = ?", id).First(tg)
	return tg
}

// 通过标签名称获取对应的数组
func GetBrowserTagsByNames(names []string, tx *gorm.DB) map[string]int64 {
	if tx == nil {
		tx = db.DB
	}
	var tgs []*BrowserTag
	tx.Model(&BrowserTag{}).Where("name in ?", names).Find(&tgs)
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
		tx.Create(&addn)
	}
	for _, v := range addn {
		mp[v.Name] = v.Id
	}

	return mp
}

// 通过id获取对应的数组
func GetBrowserTagsByIds(ids []int64) map[int64]string {
	var tgs []*BrowserTag
	db.DB.Model(&BrowserTag{}).Where("id in ?", ids).Find(&tgs)
	mp := make(map[int64]string)
	for _, v := range tgs {
		mp[v.Id] = v.Name
	}
	return mp
}

// tag标签结束----------------------------------------
// browser开始----------------------------------------

func GetBrowserById(id any) (*Browser, error) {
	b := new(Browser)
	err := db.DB.Model(&Browser{}).Where("id = ?", id).First(b).Error
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (this *Browser) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Model(&Browser{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"name":         this.Name,
				"proxy":        this.Proxy,
				"proxy_config": this.ProxyConfig,
				"lang":         this.Lang,
				"local":        this.Local,
				"timezone":     this.Timezone,
				"width":        this.Width,
				"height":       this.Height,
				"updated_at":   time.Now(),
				"ip":           this.Ip,
				"proxy_name":   this.ProxyName,
			}).Error
	} else {
		this.CreatedAt = time.Now()
		this.UpdatedAt = this.CreatedAt
		return tx.Create(this).Error
	}
}

// 删除某个Browser下的tag
func (this *Browser) RemoveMyTags(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	return tx.Where("browser_id = ?", this.Id).Delete(&BrowserToTag{}).Error
}

// 使用当前的tag标签完全替换已有标签
// 使用此方法会清空已有的tag
func (this *Browser) CoverTgs(tagsName []string, tx *gorm.DB) error {
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
		if err := tx.Create(ntag).Error; err != nil {
			return err
		}
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
	db.DB.Model(&BrowserToTag{}).Where("browser_id in ?", ids).Find(&pxtgs)

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

func (this *Browser) Open() error {
	var proxyUrl string
	if this.Proxy > 0 {
		px := proxys.GetById(this.Proxy)
		if px != nil || px.Id > 0 {
			if pc, err := px.Start(false); err == nil {
				proxyUrl = pc.Listened()
			}
		}
	}
	if proxyUrl == "" && this.ProxyConfig != "" {
		if pc, err := proxy.Client(this.ProxyConfig, "", 0); err == nil {
			if _, err := pc.Run(false); err == nil {
				proxyUrl = pc.Listened()
			}
		}
	}

	this.Bs = browser.NewBrowser(this.Lang, this.Timezone, this.Id)
	if proxyUrl != "" {
		this.Bs.SetProxy(proxyUrl, "", "")
	}
	this.Bs.SetScreen(this.Width, this.Height)
	this.Bs.SetTimezone(this.Timezone)

	if _, err := this.Bs.Run(); err != nil {
		return err
	}
	this.Opend = true
	return nil
}

func (this *Browser) Close() error {
	if bbs, ok := browser.Running.Load(this.Id); ok {
		if bs, ok := bbs.(*browser.User); ok {
			return bs.Close()
		}
	}
	return nil
}

func (this *Browser) Delete() error {
	if err := db.DB.Where("id = ?", this.Id).Delete(&Browser{}).Error; err != nil {
		return err
	}
	db.DB.Where("browser_id = ?", this.Id).Delete(&BrowserToTag{})
	if bbs, ok := browser.Running.Load(this.Id); ok {
		if bs, ok := bbs.(*browser.User); ok {
			return bs.Close()
		}
	}

	eventbus.Bus.Publish("browser-delete", "")
	return nil
}
