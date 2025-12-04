package suggestions

import (
	"time"
	"tools/runtimes/config"
	"tools/runtimes/db"

	"gorm.io/gorm"
)

type SuggCate struct {
	Id   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"index;not null"`
}
type Suggestion struct {
	Id           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Addtime      int64  `json:"addtime" gorm:"index;default:0"`               // 反馈时间
	Title        string `json:"title" gorm:"index;not null;type:varchar(40)"` // 标题
	Content      string `json:"content" gorm:"default:null"`                  // 反馈的详细内容
	CateId       int64  `json:"cate_id" gorm:"index;default:0"`               // 反馈的类别,0为未选择
	ReadTime     int64  `json:"read_time" gorm:"index;default:0"`             // 服务端阅读反馈的时间
	LastBackTime int64  `json:"last_back_time" gorm:"index;default:0"`        // 最后一次服务端反馈的时间
}
type SuggMessage struct { //对意见和建议内容进行再度讨论的内容
	Id      int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	SuggId  int64  `json:"sugg_id" gorm:"index;not null"`
	Addtime int64  `json:"addtime" gorm:"index;default:0"`
	Content string `json:"content" gorm:"not null"`                    // 内容
	Rule    int    `json:"rule" gorm:"type:tinyint(1);index;not null"` // 内容角色,0为客户端,1为服务端回答
}

func init() {
	db.DB.AutoMigrate(&SuggCate{})
	db.DB.AutoMigrate(&Suggestion{})
	db.DB.AutoMigrate(&SuggMessage{})
}

func (this *SuggCate) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	if this.Id > 0 {
		return tx.Model(&SuggCate{}).Where("id = ?", this.Id).Updates(map[string]any{
			"name": this.Name,
		}).Error
	} else {
		return tx.Create(this).Error
	}
}

func (this *Suggestion) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	if this.Id > 0 {
		return tx.Model(&Suggestion{}).Where("id = ?", this.Id).Updates(map[string]any{
			"title":          this.Title,
			"content":        this.Content,
			"cate_id":        this.CateId,
			"read_time":      this.ReadTime,
			"last_back_time": this.LastBackTime,
		}).Error
	} else {
		this.Addtime = time.Now().Unix()
		return tx.Create(this).Error
	}
}

func (this *SuggMessage) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	if this.Id > 0 {
		return tx.Model(&SuggMessage{}).Where("id = ?", this.Id).Updates(map[string]any{
			"sugg_id": this.SuggId,
			"content": this.Content,
			"rule":    this.Rule,
		}).Error
	} else {
		this.Addtime = time.Now().Unix()
		return tx.Create(this).Error
	}
}

// 将服务器返回的建议分类同步到本地
func UpSuggCateFromServer(str string) {
	if str == "" {
		return
	}
	var serverSugcts []*SuggCate
	if err := config.Json.Unmarshal([]byte(str), &serverSugcts); err == nil {
		db.DB.Migrator().DropTable(&SuggCate{})
		db.DB.AutoMigrate(&SuggCate{})
		db.DB.Create(serverSugcts)
	}
}

// 获取建议分类
func GetSuggCates() []*SuggCate {
	var sugs []*SuggCate
	db.DB.Model(&SuggCate{}).Order("id asc").Find(&sugs)
	return sugs
}
