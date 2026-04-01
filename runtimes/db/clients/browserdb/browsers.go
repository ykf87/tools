package browserdb

import (
	"errors"
	"fmt"
	"time"
	"tools/runtimes/bs"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/db/proxys"
	"tools/runtimes/eventbus"
	"tools/runtimes/mainsignal"
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
	AdminID     int64  `json:"admin_id" gorm:"index;default:0"`                       // 后台用户的登录id
	DefUrl      string `json:"def_url" gorm:"default:null"`                           // 默认打开地址
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        []string `json:"tags" gorm:"-" form:"tags"` // 标签
	// Bs          *bs.Browser `json:"-" gorm:"-" form:"-"`       // 浏览器
	Opend bool `json:"opend" gorm:"-" form:"-"` // 是否启动
	db.BaseModel
}

func init() {
	db.DB.DB().AutoMigrate(&Browser{})
	db.DB.DB().AutoMigrate(&BrowserTag{})
	db.DB.DB().AutoMigrate(&BrowserToTag{})
}

func GetBrowserById(id any) (*Browser, error) {
	b := new(Browser)
	err := db.DB.DB().Model(&Browser{}).Where("id = ?", id).First(b).Error
	if err != nil {
		return nil, err
	}
	return b, nil
}

// func (this *Browser) Save(tx *db.SQLiteWriter) error {
// 	if tx == nil {
// 		tx = db.DB
// 	}
// 	if this.Id > 0 {
// 		return tx.Write(func(txx *gorm.DB) error {
// 			return txx.Model(&Browser{}).Where("id = ?", this.Id).
// 				Updates(map[string]any{
// 					"name":         this.Name,
// 					"proxy":        this.Proxy,
// 					"proxy_config": this.ProxyConfig,
// 					"lang":         this.Lang,
// 					"local":        this.Local,
// 					"timezone":     this.Timezone,
// 					"width":        this.Width,
// 					"height":       this.Height,
// 					"updated_at":   time.Now(),
// 					"ip":           this.Ip,
// 					"proxy_name":   this.ProxyName,
// 				}).Error
// 		})
// 	} else {
// 		this.CreatedAt = time.Now()
// 		this.UpdatedAt = this.CreatedAt
// 		return tx.Write(func(txx *gorm.DB) error {
// 			return txx.Create(this).Error
// 		})
// 	}
// }

func (this *Browser) Open(opt *bs.Options) error {
	if opt == nil {
		opt = &bs.Options{
			ID:       this.Id,
			Width:    this.Width,
			Height:   this.Height,
			Language: this.Lang,
			Timezone: this.Timezone,
			Url:      this.DefUrl,
			Show:     true,
		}
	}
	browser, err := bs.BsManager.New(this.Id, opt, true)

	if browser.IsArrive() {
		return fmt.Errorf("浏览器已经打开")
	}

	if opt.Pc == nil {
		if this.Proxy > 0 {
			px := proxys.GetById(this.Proxy)
			if px != nil && px.Id > 0 {
				if pc, err := px.Start(false); err == nil {
					opt.Pc = pc
				}
			}
		} else if this.ProxyConfig != "" {
			if pc, err := proxy.Client(this.ProxyConfig, "", 0, ""); err == nil {
				if _, err := pc.Run(false); err == nil {
					opt.Pc = pc
				}
			}
		}
	}
	if this.DefUrl != "" {
		opt.Url = this.DefUrl
	}

	opt.Width = this.Width
	opt.Height = this.Height
	opt.Language = this.Lang
	opt.Timezone = this.Timezone
	opt.ID = this.Id
	opt.Show = true

	bbs, err := bs.BsManager.New(this.Id, opt, true)
	if err != nil {
		return err
	}
	// this.Bs = bbs

	// this.Bs.OnClosed(func() {
	// 	this.Opend = false
	// 	eventbus.Bus.Publish("browser-close", this)
	// })
	// this.Bs.OnConsole(func(args []*runtime.RemoteObject) {
	// 	fmt.Println(args, "args-----")
	// })

	// this.Bs.OnURLChange(func(url string) {
	// 	if this.Bs.Opts.JsStr != "" {
	// 		this.Bs.RunJs(this.Bs.Opts.JsStr)
	// 	}
	// })

	if err := bbs.OpenBrowser(); err != nil {
		return err
	}
	// this.Opend = true
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

// 批量删除
func BatchDelete(ids []int64) error {

	var bs []*Browser
	db.DB.DB().Where("id in ?", ids).Find(&bs)

	for _, v := range bs {
		v.Close(true)
	}
	return db.DB.Write(func(tx *gorm.DB) error {
		if err := tx.Where("id in ?", ids).Delete(&Browser{}).Error; err != nil {
			return err
		}
		return tx.Where("browser_id in ?", ids).Delete(&BrowserToTag{}).Error
	})
}

func (this *Browser) Delete() error {
	if err := db.DB.Write(func(tx *gorm.DB) error {
		return tx.Where("id = ?", this.Id).Delete(&Browser{}).Error
	}); err != nil {
		return err
	}

	if err := db.DB.Write(func(txx *gorm.DB) error {
		if err := txx.Where("browser_id = ?", this.Id).Delete(&BrowserToTag{}).Error; err != nil {
			return err
		}
		if err := this.Close(true); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	// tx := db.DB.Begin()
	// tx.Where("browser_id = ?", this.Id).Delete(&BrowserToTag{})
	// if err := this.Close(false); err != nil {
	// 	tx.Rollback()
	// 	return err
	// }
	// tx.Commit()

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
	b, err := bs.BsManager.GetBrowser(this.Id)
	if err == nil {
		return b
	}
	return nil
}

// 生成浏览器代理的proxy_config
func (this *Browser) GenProxyConfig() (*proxy.ProxyConfig, error) {
	if this.Proxy > 0 {
		return proxys.GetProxyConfigByID(this.Proxy)
	} else if this.ProxyConfig != "" {
		return proxy.Client(this.ProxyConfig, "", 0, "")
	}
	return nil, errors.New("No Proxy")
}

// 通过id构建浏览器的opt
// 无论如何都会创建
func GenBrowserOpt(id int64, isshow bool) *bs.Options {
	var width int
	var height int
	if w, ok := config.AdminWidthAndHeight.Load("width"); ok {
		width, _ = w.(int)
	}
	if h, ok := config.AdminWidthAndHeight.Load("height"); ok {
		height, _ = h.(int)
	}
	opt := &bs.Options{
		Width:    width,
		Height:   height,
		Show:     isshow,
		ID:       0,
		Url:      "",
		JsStr:    "",
		Timezone: "",
		Language: "",
		Ctx:      mainsignal.MainCtx,
		Timeout:  time.Second * 60,
		Pc:       nil,
		Temp:     true,
	}

	row, err := GetBrowserById(id)
	if err != nil {
		return opt
	}

	opt.Width = row.Width
	opt.Height = row.Height
	opt.Language = row.Lang
	opt.Timezone = row.Timezone
	opt.ID = row.Id
	opt.Temp = false
	if row.Proxy > 0 {
		pp := proxys.GetById(row.Proxy)
		if pp != nil {
			go func() {
				if ppp, err := pp.Start(false); err == nil {
					opt.Pc = ppp
				}
			}()
		}
	} else if row.ProxyConfig != "" {
		if ppp, err := proxy.Client(row.ProxyConfig, "", 0, ""); err != nil {
			opt.Pc = ppp
		}
	}

	return opt
}
