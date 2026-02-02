package httpurl

import "tools/runtimes/db"

type HttpUrl struct {
	ID      int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name    string `json:"name" gorm:"index"`     // 自己命名的名称
	Url     string `json:"url" gorm:"primaryKey"` // http发起的url, 从此处开始为 device_type = 2的参数
	Method  string `json:"method"`                // http 请求方式
	Data    string `json:"data"`                  // 如果是post发起的，携带数据
	Cookies string `json:"cookies"`               // 使用的cookie
	Headers string `json:"headers"`               // 携带的头部信息
}

func GetHttpUrlByID(id any) (*HttpUrl, error) {
	var hu *HttpUrl
	if err := db.DB.Model(&HttpUrl{}).Where("id = ?", id).First(hu).Error; err != nil {
		return nil, err
	}
	return hu, nil
}
