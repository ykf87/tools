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
	Password string `json:"password" gorm:"default:null"`
	CreateAt string `json:"create_at" gorm:"index"`
	UpdateAt string `json:"update_at" gorm:"index"`
	Status   int    `json:"status" gorm:"type:tinyint(1);"`
	Timer    int64  `json:"timer" gorm:"type:int(64)"`
	Jwt      string `json:"-" gorm:"-"`
}

func init() {
	db.DB.AutoMigrate(&Admin{})
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

	if err := funcs.VerifyPassword(adm.Password, password); err != nil {
		return nil, err
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

func (this *Admin) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}
	if this.Id > 0 {
		return tx.Model(&Admin{}).Where("id = ?", this.Id).
			Updates(Admin{
				Name:     this.Name,
				Password: this.Password,
				UpdateAt: time.Now().Format("2006-01-02 15:04:05"),
				Timer:    this.Timer,
				Status:   this.Status,
			}).Error
	} else {
		this.CreateAt = time.Now().Format("2006-01-02 15:04:05")
		return tx.Create(this).Error
	}
}
