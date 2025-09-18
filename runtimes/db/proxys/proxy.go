package proxys

type Proxy struct {
	Id        int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string `json:"name" gorm:"default:null;index"`     // 名称
	Remark    string `json:"remark" gorm:"default:null;index"`   // 备注
	Local     string `json:"local" gorm:"default:null;index"`    // 地区
	Subscribe int64  `json:"subscribe" gorm:"index;default:0"`   // 订阅的id,订阅的代理额外管理
	Port      int    `json:"port" gorm:"index;default:0"`        // 指定的端口,不存在则随机使用空余端口
	Config    string `json:"config" gorm:"not null;"`            // 代理信息,可以是vmess,vless等,也可以是http代理等
	Username  string `json:"username" gorm:"default:null;index"` // 有些http代理等需要用户名
	Password  string `json:"password" gorm:"default:null"`       // 对应的密码
	Transfer  string `json:"transfer"`                           // 有些代理需要中转,无法直连.目的是解决有的好的ip在国外无法通过国内直连
	AutoRun   int    `json:"auto_run" gorm:"default:0;index"`    // 系统启动跟随启动
}
