package proxys

import (
	"strconv"
	"tools/runtimes/db"
	"tools/runtimes/proxy"

	"gorm.io/gorm"
)

type Proxy struct { // 如果有修改字段,需要更新Save方法
	Id        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string `json:"name" gorm:"default:null;index"`     // 名称
	Remark    string `json:"remark" gorm:"default:null;index"`   // 备注
	Local     string `json:"local" gorm:"default:null;index"`    // 地区
	Subscribe int64  `json:"subscribe" gorm:"index;default:0"`   // 订阅的id,订阅的代理额外管理
	Port      int    `json:"port" gorm:"index;default:0"`        // 指定的端口,不存在则随机使用空余端口
	Config    string `json:"config" gorm:"not null;"`            // 代理信息,可以是vmess,vless等,也可以是http代理等
	Username  string `json:"username" gorm:"default:null;index"` // 有些http代理等需要用户名
	Password  string `json:"password" gorm:"default:null"`       // 对应的密码
	Transfer  string `json:"transfer" gorm:"default:null"`       // 有些代理需要中转,无法直连.目的是解决有的好的ip在国外无法通过国内直连,可以是proxy的id或者具体配置
	AutoRun   int    `json:"auto_run" gorm:"default:0;index"`    // 系统启动跟随启动
}

func init() {
	db.DB.AutoMigrate(&Proxy{})

	//随系统启动的代理
	var proxys []*Proxy
	db.DB.Model(&Proxy{}).Where("auto_run = 1").Find(&proxys)
	for _, v := range proxys {
		var trans []string
		if v.Transfer != "" {
			if vid, err := strconv.Atoi(v.Transfer); err == nil {
				row := new(Proxy)
				if err := db.DB.Model(&Proxy{}).Where("id = ?", vid).First(row).Error; err == nil {
					if row.Config != "" {
						trans = append(trans, row.Config)
					}
				}
			} else {
				trans = append(trans, v.Transfer)
			}
		}
		proxy.Run(v.Config, "", v.Port, true, trans...)
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
