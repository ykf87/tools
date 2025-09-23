package proxys

import (
	"tools/runtimes/db"
	"tools/runtimes/db/tag"
	"tools/runtimes/proxy"

	"gorm.io/gorm"
)

type ProxyTag struct {
	ProxyId int64 `json:"pid" gorm:"primaryKey;not null"` // 代理id
	TagId   int64 `json:"tid" gorm:"primaryKey;not null"` // 标签id
}

type Proxy struct { // 如果有修改字段,需要更新Save方法
	Id        int64    `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name      string   `json:"name" gorm:"default:null;index" form:"name"`         // 名称
	Remark    string   `json:"remark" gorm:"default:null;index" form:"remark"`     // 备注
	Local     string   `json:"local" gorm:"default:null;index"`                    // 地区
	Subscribe int64    `json:"subscribe" gorm:"index;default:0" form:"subscribe"`  // 订阅的id,订阅的代理额外管理
	Port      int      `json:"port" gorm:"index;default:0" form:"port"`            // 指定的端口,不存在则随机使用空余端口
	Config    string   `json:"config" gorm:"not null;" form:"config"`              // 代理信息,可以是vmess,vless等,也可以是http代理等
	Username  string   `json:"username" gorm:"default:null;index" form:"username"` // 有些http代理等需要用户名
	Password  string   `json:"password" gorm:"default:null" form:"password"`       // 对应的密码
	Transfer  string   `json:"transfer" gorm:"default:null" form:"transfer"`       // 有些代理需要中转,无法直连.目的是解决有的好的ip在国外无法通过国内直连,可以是proxy的id或者具体配置
	AutoRun   int      `json:"auto_run" gorm:"default:0;index" form:"auto_run"`    // 系统启动跟随启动
	Tags      []string `json:"tags" gorm:"-" form:"tags"`                          // 标签列表,不写入数据库,仅在添加和修改时使用
	// Gid       int64    `json:"gid" gorm:"default:0;index"`         // 分组
}

func init() {
	db.DB.AutoMigrate(&Proxy{})

	//随系统启动的代理
	var proxys []*Proxy
	db.DB.Model(&Proxy{}).Where("auto_run = 1").Find(&proxys)
	for _, v := range proxys {
		v.Start(true)
		if v.Local == "" {
			if local, err := proxy.GetLocal(v.Config, v.Transfer); err == nil {
				v.Local = local
				v.Save(nil)
			}
		}
	}
}

// 保存
func (this *Proxy) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Model(&Proxy{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"name":      this.Name,
				"remark":    this.Remark,
				"local":     this.Local,
				"subscribe": this.Subscribe,
				"config":    this.Config,
				"username":  this.Username,
				"password":  this.Password,
				"transfer":  this.Transfer,
				"auto_run":  this.AutoRun,
			}).Error
	} else {
		return tx.Create(this).Error
	}
}

// 启动配置的代理
// keep 是否守护代理
func (this *Proxy) Start(keep bool) (*proxy.ProxyConfig, error) {
	p, err := proxy.Client(this.Config, "", this.Port, this.Transfer)
	if err != nil {
		return nil, err
	}

	return p.Run(keep)
}

// 通过id获取代理
func GetById(id any) *Proxy {
	var px *Proxy
	db.DB.Where("id = ?", id).First(px)
	return px
}

// 通过端口获取代理
func GetByPort(port any) *Proxy {
	var px *Proxy
	db.DB.Where("port = ?", port).First(px)
	return px
}

// 使用当前的tag标签完全替换已有标签
// 使用此方法会清空已有的tag
func (this *Proxy) CoverTgs(tagsName []string) error {
	if err := db.DB.Where("pid = ?", this.Id).Delete(&ProxyTag{}).Error; err != nil {
		return err
	}

	var tgs []*tag.Tag
	db.DB.Model(&tag.Tag{}).Where("name in ?", tagsName).Find(&tgs)
	var pts []*ProxyTag
	for _, v := range tgs {
		pts = append(pts, &ProxyTag{ProxyId: this.Id, TagId: v.Id})
	}
	if len(pts) > 0 {
		return db.DB.Create(pts).Error
	}
	return nil
}

// 返回当前代理是否已启动
func (this *Proxy) IsStart() *proxy.ProxyConfig {
	return proxy.IsRuning(this.Config)
}

// 删除当前的
func (this *Proxy) Remove() error {
	return db.DB.Where("id = ?", this.Id).Delete(&Proxy{}).Error
}
