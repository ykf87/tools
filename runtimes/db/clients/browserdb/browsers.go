package browserdb

import (
	"fmt"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/db"
	"tools/runtimes/db/proxys"
	"tools/runtimes/eventbus"
	"tools/runtimes/proxy"

	"github.com/chromedp/cdproto/runtime"
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
	AdminID     int64  `json:"admin_id" gorm:"index;default:0"`                       // 后台用户的登录id
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        []string    `json:"tags" gorm:"-" form:"tags"` // 标签
	Bs          *bs.Browser `json:"-" gorm:"-" form:"-"`       // 浏览器
	Opend       bool        `json:"opend" gorm:"-" form:"-"`   // 是否启动
}

func init() {
	db.DB.AutoMigrate(&Browser{})
	db.DB.AutoMigrate(&BrowserTag{})
	db.DB.AutoMigrate(&BrowserToTag{})
}

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
			Updates(map[string]any{
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

func (this *Browser) Open() error {
	if this.Opend == true {
		return nil
	}
	var proxyUrl string
	if this.Proxy > 0 {
		px := proxys.GetById(this.Proxy)
		if px != nil && px.Id > 0 {
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

	bbs, err := bs.BsManager.New(this.Id, bs.Options{
		Proxy:    proxyUrl,
		Width:    this.Width,
		Height:   this.Height,
		Language: this.Lang,
		Timezone: this.Timezone,
	}, true)
	if err != nil {
		return err
	}
	this.Bs = bbs

	this.Bs.OnClosed(func() {
		this.Opend = false
		eventbus.Bus.Publish("browser-close", this)
	})
	this.Bs.OnConsole(func(args []*runtime.RemoteObject) {
		fmt.Println(args, "args-----")
	})

	this.Bs.OnURLChange(func(url string) {
		if this.Bs.Opts.JsStr != "" {
			this.Bs.RunJs(this.Bs.Opts.JsStr)
		}
	})

	if err := this.Bs.OpenBrowser(); err != nil {
		return err
	}
	this.Opend = true
	return nil

}

// rmv是否彻底删除
func (this *Browser) Close(rmv bool) error {
	if err := bs.BsManager.Close(this.Id); err != nil {
		return err
	}

	if rmv == true {
		bs.BsManager.Remove(this.Id)
	}

	eventbus.Bus.Publish("browser-close", this)
	return nil
}

func (this *Browser) Delete() error {
	if err := db.DB.Where("id = ?", this.Id).Delete(&Browser{}).Error; err != nil {
		return err
	}

	tx := db.DB.Begin()
	tx.Where("browser_id = ?", this.Id).Delete(&BrowserToTag{})
	if err := this.Close(false); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	eventbus.Bus.Publish("browser-delete", "")
	return nil
}

// -------------------------  以下是一个interface的实现  -------------------------------
// 获取浏览器的id
func (this *Browser) GetId() int64 {
	return this.Id
}

// 获取客户端
func (this *Browser) GetClient() *bs.Browser {
	return this.Bs
}
