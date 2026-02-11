package proxys

import (
	"fmt"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"
	"tools/runtimes/funcs"
	"tools/runtimes/logs"
	"tools/runtimes/services"

	"gorm.io/gorm"
)

type Subscribe struct {
	Id      int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name    string `json:"name" gorm:"not null;uniqueIndex;"` // 自己备注的名称,必填且不重复
	Url     string `json:"url" gorm:"not null"`               // 订阅地址,不能重复订阅
	UrlMd5  string `json:"url_md5" gorm:"not null;uniqueIndex"`
	Addtime int64  `json:"addtime" gorm:"index;default:0"`
	Total   int    `json:"total" gorm:"index;default:0"` // 订阅地址下的代理数量
}

var DefSubUrl string

func init() {
	db.DB.DB().AutoMigrate(&Subscribe{})
	DefSubUrl = fmt.Sprint(config.SERVERDOMAIN, "subscription")
	if _, err := AddNewSub(DefSubUrl, "系统赠送"); err != nil {
		logs.Error(err.Error())
	}
}

func AddNewSub(urlStr, name string) (*Subscribe, error) {
	urlmd5 := funcs.Md5String(urlStr)
	subrow := new(Subscribe)
	db.DB.DB().Model(&Subscribe{}).Where("url_md5 = ?", urlmd5).First(subrow)
	if subrow.Id < 1 { // 添加默认订阅地址和订阅代理
		subrow := &Subscribe{
			Name:   name,
			Url:    DefSubUrl,
			UrlMd5: urlmd5,
			Total:  0,
		}
		if err := subrow.Save(nil); err != nil {
			return nil, err
		}

		if _, err := subrow.GetFromServer(); err != nil {
			return nil, err
		}
	}
	return subrow, nil
}

func (this *Subscribe) GetFromServer() ([]*Proxy, error) {
	resp, err := services.GerProxySub(this.Url)
	if err != nil {
		return nil, err
	}

	db.DB.Write(func(tx *gorm.DB) error {
		return tx.Where("subscribe = ?", this.Id).Delete(&Proxy{}).Error
	})

	// var md5s []string
	//  md5Resp := make(map[string]*services.SubResp)
	// for _, v := range resp{
	// 	// md5s = append(md5s, v.ConfigMd5)
	// 	md5Resp[v.ConfigMd5] = v
	// }

	// var proxys []*Proxy
	// db.DB.Model(&Proxy{}).Where("subscribe = ?", this.Id).Updates(map[string]interface{}{
	// 	"deleted": 1,
	// })

	// for _, v := range proxys{
	// 	if rrv, ok := md5Resp[v.ConfigMd5]; ok{
	// 		v.AutoRun = rrv.AutoRun
	// 		v.Private = rrv.Private
	// 		v.SubName = this.Name
	// 		v.Subscribe = this.Id
	// 		v.Save(nil)
	// 		delete(md5Resp, v.ConfigMd5)
	// 	}
	// }

	var newProxys []*Proxy
	for _, v := range resp {
		pro := new(Proxy)
		pro.AutoRun = v.AutoRun
		pro.Config = v.Config
		pro.ConfigMd5 = v.ConfigMd5
		pro.Encrypt = 1
		pro.Ip = v.Ip
		pro.Lang = v.Lang
		pro.Local = v.Local
		pro.Name = v.Name
		pro.Password = v.Password
		pro.Port = v.Port
		pro.Private = v.Private
		pro.Remark = v.Remark
		pro.SubName = this.Name
		pro.Subscribe = this.Id
		pro.Timezone = v.Timezone
		pro.Transfer = v.Transfer
		pro.Username = v.Username
		pro.Addtime = time.Now().Unix()
		newProxys = append(newProxys, pro)
	}
	if len(newProxys) > 0 {
		db.DB.Write(func(tx *gorm.DB) error {
			return tx.Model(&Proxy{}).Create(newProxys).Error
		})
	}
	this.Total = len(resp)
	this.Save(nil)

	return newProxys, nil
}

func (this *Subscribe) Save(tx *db.SQLiteWriter) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Write(func(txx *gorm.DB) error {
			return txx.Model(&Subscribe{}).Where("id = ?", this.Id).
				Updates(map[string]interface{}{
					"name":    this.Name,
					"url":     this.Url,
					"url_md5": this.UrlMd5,
					"total":   this.Total,
				}).Error
		})
	} else {
		this.Addtime = time.Now().Unix()
		return tx.Write(func(txx *gorm.DB) error {
			return txx.Create(this).Error
		})
	}
}
