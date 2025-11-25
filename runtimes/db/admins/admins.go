package admins

import (
	"fmt"
	"time"
	"tools/runtimes/db"
	"tools/runtimes/funcs"
	"tools/runtimes/i18n"

	"gorm.io/gorm"
)

type Admin struct {
	Id       int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Account  string `json:"account" gorm:"uniqueIndex;not null"`
	Name     string `json:"name" gorm:"index"`
	Password string `json:"password" gorm:"default:null" parse:"-"`
	CreateAt string `json:"create_at" gorm:"index"`
	UpdateAt string `json:"update_at" gorm:"index"`
	Status   int    `json:"status" gorm:"type:tinyint(1);"`
	Timer    int64  `json:"timer" gorm:"type:int(64)" parse:"-"`
	Main     int    `json:"main" gorm:"type:tinyint(1);index"`
	Jwt      string `json:"-" gorm:"-" parse:"-"`
	Group string `json:"group" gorm:"default:null"`
}

func init() {
	db.DB.AutoMigrate(&Admin{})

	var reslen int64
	db.DB.Model(&Admin{}).Count(&reslen)
	if reslen < 1 {
		adm := &Admin{
			Account:  "admin",
			Name:     "Super",
			Password: "",
			Status:   1,
			Main:     1,
			Group: "admin",
		}
		adm.Save(nil)
	}
}

// 登录
func Login(account, password string) (*Admin, error) {
	adm := new(Admin)
	if err := db.DB.Model(&Admin{}).Where("account = ?", account).First(adm).Error; err != nil {
		return nil, err
	}

	if adm == nil || adm.Id < 1 {
		return nil, fmt.Errorf(i18n.T("%s cannot found", account))
	}

	switch adm.Status {
	case 1:
		break
	case 0:
		return nil, fmt.Errorf(i18n.T("Account %s is not activated", account))
	case -1:
		return nil, fmt.Errorf(i18n.T("Account %s is illegal", account))
	default:
		return nil, fmt.Errorf(i18n.T("Account %s status error", account))
	}

	if adm.Password != "" {
		if err := funcs.VerifyPassword(adm.Password, password); err != nil {
			return nil, fmt.Errorf(i18n.T("Password error"))
		}
	} else if password != "" {
		adm.Password, _ = funcs.GenPassword(password, 0)
	}

	adm.Timer = adm.Timer + 1
	adm.Save(nil)

	jwt, err := adm.GenJwt()
	if err != nil {
		return nil, err
	}
	adm.Jwt = jwt

	return adm, nil
}

// 保存
func (this *Admin) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Model(&Admin{}).Where("id = ?", this.Id).
			Updates(map[string]interface{}{
				"name":      this.Name,
				"password":  this.Password,
				"update_at": time.Now().Format("2006-01-02 15:04:05"),
				"timer":     this.Timer,
				"status":    this.Status,
			}).Error
	} else {
		this.CreateAt = time.Now().Format("2006-01-02 15:04:05")
		this.UpdateAt = this.CreateAt
		return tx.Create(this).Error
	}
}

// 获取用户列表
func AdminList(page, limit int, q, bykey, by string) ([]*Admin, int64) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	var adms []*Admin
	model := db.DB.Model(&Admin{})
	if q != "" {
		qs := fmt.Sprintf("%%%s%%", q)
		model = model.Where("account LIKE ? OR name LIKE ?", qs, qs)
	}
	if bykey == "" {
		bykey = "id"
	}
	switch by {
	case "asc":
		by = "ASC"
	default:
		by = "DESC"
	}

	var totals int64
	model.Count(&totals)

	model.Order(fmt.Sprintf("%s %s", bykey, by)).Offset((page - 1) * limit).Limit(limit).Find(&adms)
	return adms, totals
}

// 使用id获取用户
func GetAdminFromId(id any) *Admin {
	adm := new(Admin)
	db.DB.Model(&Admin{}).Where("id = ?", id).First(adm)
	return adm
}

// 根据id删除用户
func DeleteAdminById(id any) error {
	adm := GetAdminFromId(id)
	if adm == nil || adm.Id < 1 {
		return fmt.Errorf(i18n.T("Account not found"))
	}

	if adm.Main == 1 {
		return fmt.Errorf(i18n.T("Super administrators cannot delete"))
	}

	return db.DB.Where("id = ?", id).Delete(&Admin{}).Error
}
