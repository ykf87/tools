package proxys

type Proxy struct {
	Id     int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name   string `json:"name" gorm:"default:null;index"`   // 名称
	Remark string `json:"remark" gorm:"default:null;index"` // 备注
	Local  string `json:"local" gorm:"default:null;index"`  // 地区
	Port   int    `json:"port" gorm:"index;default:0"`      // 指定的端口,不存在则随机使用空余端口
}
