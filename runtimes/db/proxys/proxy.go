package proxys

import (
	"errors"
	"fmt"
	"time"
	"tools/runtimes/aess"
	"tools/runtimes/db"
	"tools/runtimes/db/tag"
	"tools/runtimes/eventbus"
	"tools/runtimes/proxy"

	"gorm.io/gorm"
)

type ProxyTag struct {
	ProxyId int64 `json:"pid" gorm:"primaryKey;not null"` // 代理id
	TagId   int64 `json:"tid" gorm:"primaryKey;not null"` // 标签id
}

type Proxy struct { // 如果有修改字段,需要更新Save方法
	Id         int64    `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name       string   `json:"name" gorm:"default:null;index" form:"name"`         // 名称
	Remark     string   `json:"remark" gorm:"default:null;index" form:"remark"`     // 备注
	Local      string   `json:"local" gorm:"default:null;index"`                    // 地区
	Ip         string   `json:"ip" gorm:"default:null;index;"`                      // 代理的ip地址
	Timezone   string   `json:"timezone" gorm:"default:null;"`                      // 代理的时区
	Lang       string   `json:"lang" gorm:"default:null"`                           // 代理所在地区使用的语言
	Subscribe  int64    `json:"subscribe" gorm:"index;default:0" form:"subscribe"`  // 订阅的id,订阅的代理额外管理
	SubName    string   `json:"sub_name" gorm:"default:null"`                       // 订阅的名称
	Port       int      `json:"port" gorm:"index;default:0" form:"port"`            // 指定的端口,不存在则随机使用空余端口
	Config     string   `json:"config" gorm:"not null;" form:"config"`              // 代理信息,可以是vmess,vless等,也可以是http代理等
	ConfigMd5  string   `json:"config_md5" gorm:"index;not null"`                   // 配置的md5,用于去重
	Username   string   `json:"username" gorm:"default:null;index" form:"username"` // 有些http代理等需要用户名
	Password   string   `json:"password" gorm:"default:null" form:"password"`       // 对应的密码
	Transfer   string   `json:"transfer" gorm:"default:null" form:"transfer"`       // 有些代理需要中转,无法直连.目的是解决有的好的ip在国外无法通过国内直连,可以是proxy的id或者具体配置
	AutoRun    int      `json:"auto_run" gorm:"default:0;index" form:"auto_run"`    // 系统启动跟随启动
	Encrypt    int      `json:"encrypt" gorm:"index;type:tinyint(1);default:0"`     // 配置是否加密,服务端拿到的是加密的,防止被用于别处
	Private    int      `json:"private" gorm:"index;default:0;type:tinyint(1)"`     // 是否是私有的
	Deleted    int      `json:"deleted" gorm:"index;type:tinyint(1);default:0"`     // 是否无效
	Addtime    int64    `json:"addtime" gorm:"index;default:0"`                     // 添加事件
	Tags       []string `json:"tags" gorm:"-" form:"tags"`                          // 标签列表,不写入数据库,仅在添加和修改时使用
	IsRuning   int      `json:"is_runing" gorm:"-" form:"-"`                        // 是否启动
	ListerAddr string   `json:"lister_addr" gorm:"-" form:"-"`                      // 监听地址
	// Ping       string   `json:"ping" gorm:"-" form:"-"`                             // 测速结果
	// Gid       int64    `json:"gid" gorm:"default:0;index"`         // 分组
}

// 延迟
type PingResp struct {
	UID  int64           `json:"uid"`
	Ping map[int64]int64 `json:"ping"`
}

func init() {
	db.DB.AutoMigrate(&Proxy{})
	db.DB.AutoMigrate(&ProxyTag{})

	//随系统启动的代理
	go func() {
		var proxys []*Proxy
		db.DB.Model(&Proxy{}).Where("auto_run = 1").Find(&proxys)
		for _, v := range proxys {
			v.Start(true)
			if v.Local == "" {
				if local, err := proxy.GetLocal(v.GetConfig(), v.GetTransfer()); err == nil {
					v.Local = local.Iso
					v.Ip = local.Ip
					v.Timezone = local.Timezone
					v.Lang = local.Lang
					v.Save(nil)
				}
			}
		}
	}()
}

// 保存
func (this *Proxy) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		err := tx.Model(&Proxy{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"name":       this.Name,
				"remark":     this.Remark,
				"local":      this.Local,
				"ip":         this.Ip,
				"lang":       this.Lang,
				"timezone":   this.Timezone,
				"subscribe":  this.Subscribe,
				"config":     this.Config,
				"config_md5": this.ConfigMd5,
				"username":   this.Username,
				"password":   this.Password,
				"transfer":   this.Transfer,
				"auto_run":   this.AutoRun,
				"port":       this.Port,
				"private":    this.Private,
				"encrypt":    this.Encrypt,
				"deleted":    this.Deleted,
			}).Error
		if err != nil {
			return err
		}
		if this.Deleted == 1 {
			this.Stop(true)
		}
		eventbus.Bus.Publish("proxy_change", this)
		return nil
	} else {
		this.Addtime = time.Now().Unix()
		return tx.Create(this).Error
	}
}

// 返回正确的config代理内容
func (this *Proxy) GetConfig() string {
	if this.Encrypt == 1 {
		return aess.AesDecryptCBC(this.Config)
	}
	return this.Config
}

// 返回正确的transfer代理内容
func (this *Proxy) GetTransfer() string {
	if this.Encrypt == 1 {
		return aess.AesDecryptCBC(this.Transfer)
	}
	return this.Transfer
}

// 启动配置的代理
// keep 是否守护代理
func (this *Proxy) Start(keep bool) (*proxy.ProxyConfig, error) {
	if this.Deleted == 1 {
		return nil, errors.New("代理已废弃或删除")
	}
	p, err := proxy.Client(this.GetConfig(), "", this.Port, this.GetTransfer())
	if err != nil {
		return nil, err
	}

	return p.Run(keep)
}

// 通过代理id获得 *ProxyConfig
func GetProxyConfigByID(id int64) (*proxy.ProxyConfig, error) {
	p := GetById(id)
	if p == nil || p.Id < 1 {
		return nil, fmt.Errorf("代理不存在")
	}

	pc, err := proxy.Client(p.GetConfig(), "", p.Port, p.GetTransfer())
	if err != nil {
		return nil, err
	}

	return pc, nil
}

// 停止配置的代理
// enforce 是否强制关闭
func (this *Proxy) Stop(enforce bool) error {
	pc, err := proxy.Client(this.GetConfig(), "", 0)
	if err != nil {
		return err
	}
	return pc.Close(enforce)
}

// 通过id获取代理
func GetById(id any) *Proxy {
	px := new(Proxy)
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
func (this *Proxy) CoverTgs(tagsName []string, tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	if err := this.RemoveMyTags(tx); err != nil {
		return err
	}

	mp := tag.GetTagsByNames(tagsName, tx)

	var ntag []ProxyTag
	for _, tagg := range tagsName {
		if tid, ok := mp[tagg]; ok {
			ntag = append(ntag, ProxyTag{ProxyId: this.Id, TagId: tid})
		}
	}

	if len(ntag) > 0 {
		if err := tx.Create(ntag).Error; err != nil {
			return err
		}
	}

	// var tgs []*tag.Tag
	// db.DB.Model(&tag.Tag{}).Where("name in ?", tagsName).Find(&tgs)
	// var pts []*ProxyTag
	// for _, v := range tgs {
	// 	pts = append(pts, &ProxyTag{ProxyId: this.Id, TagId: v.Id})
	// }
	// if len(pts) > 0 {
	// 	return db.DB.Create(pts).Error
	// }
	return nil
}

// 返回当前代理是否已启动
func (this *Proxy) IsStart() *proxy.ProxyConfig {
	return proxy.IsRuning(this.GetConfig())
}

// 删除当前的
func (this *Proxy) Remove() error {
	return db.DB.Where("id = ?", this.Id).Delete(&Proxy{}).Error
}

// 删除某个proxy下的tag
func (this *Proxy) RemoveMyTags(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	return tx.Where("proxy_id = ?", this.Id).Delete(&ProxyTag{}).Error
}

type Psc []*Proxy

func SetTags(pcs []*Proxy) {
	if len(pcs) < 1 {
		return
	}
	var ids []int64
	for _, v := range pcs {
		ids = append(ids, v.Id)
	}

	var pxtgs []*ProxyTag
	db.DB.Model(&ProxyTag{}).Where("proxy_id in ?", ids).Find(&pxtgs)

	var tagids []int64
	pcMap := make(map[int64][]int64)
	for _, v := range pxtgs {
		tagids = append(tagids, v.TagId)
		pcMap[v.ProxyId] = append(pcMap[v.ProxyId], v.TagId)
	}

	ttggs := tag.GetTagsByIds(tagids)

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
