// 执行的js脚本,浏览器和手机端都是使用js来执行
package jses

import (
	"fmt"
	"strings"
	"time"
	"tools/runtimes/aess"
	"tools/runtimes/db"
	"tools/runtimes/funcs"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// js脚本表,content是脚本的内容
// replace_prev 默认是 <<
// replace_end 默认是 >>
type Js struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Code        string     `json:"code" gorm:"uniqueIndex;not null"`              // 唯一标识符
	Name        string     `json:"name" gorm:"index"`                             // 名称
	Tp          int        `json:"tp" gorm:"type:tinyint(1);index;default:0"`     // 脚本类型,适用什么设备的, 0-代表web端  1-代表手机端autos
	IsSys       int        `json:"is_sys" gorm:"type:tinyint(1);index;default:0"` // 是否是从服务端获取的脚本,如果从服务器获取的脚本,将使用aes加密,1为系统获取, 0为用户自写
	AdminID     int64      `json:"admin_id" gorm:"index;default:0"`               // 管理员id, 如果是系统的则为0,如果是用户自己写的,则对应用户的id
	Content     string     `json:"content" gorm:"not null;type:longtext"`         // 执行的脚本
	ReplacePrev string     `json:"replace_prev"`                                  // 变量替换前缀
	ReplaceEnd  string     `json:"replace_end"`                                   // 变量替换后缀
	Icon        string     `json:"icon"`                                          // 此js的图标
	Addtime     int64      `json:"addtime" gorm:"index;default:0"`                // 添加时间
	Def         string     `json:"def" gorm:"default:null"`                       // 默认网址或者app
	Tags        []string   `json:"tags" gorm:"-"`                                 // 标签
	Params      []*JsParam `json:"params" gorm:"-"`                               // 参数
}

type RplsContent struct {
	Code  string `json:"code_name"`
	Value string `json:"value"`
	// Api
}

func init() {
	db.DB.AutoMigrate(&Js{})
	db.DB.AutoMigrate(&JsParam{})
}

func (this *Js) Save(tx *gorm.DB) error {
	if tx == nil {
		tx = db.DB
	}

	if this.Name == "" {
		return fmt.Errorf("请填写脚本标题")
	}

	// defer NotifyTaskChanged(this.ID)
	if this.ID > 0 {
		return tx.Model(&Js{}).Where("id = ?", this.ID).
			Updates(map[string]any{
				"name":         this.Name,
				"code":         this.Code,
				"content":      this.Content,
				"replace_prev": this.ReplacePrev,
				"replace_end":  this.ReplaceEnd,
				"icon":         this.Icon,
				"admin_id":     this.AdminID,
			}).Error
	} else {
		if this.Addtime < 1 {
			this.Addtime = time.Now().Unix()
		}
		// this.IsSys = 1
		err := tx.Create(this).Error
		return err
	}
}

// 获取js的内容
func (this *Js) GetContent(taskParams map[string]any) string {
	params := this.GetParams()
	if len(params) < 1 {
		return this.Content
	}
	str := this.Content
	if this.IsSys == 1 {
		str = aess.AesDecryptCBC(str)
	}

	prev := "{{"
	end := "}}"
	if this.ReplacePrev != "" {
		prev = this.ReplacePrev
	}
	if this.ReplaceEnd != "" {
		end = this.ReplaceEnd
	}

	for _, v := range params {
		val, ok := taskParams[v.CodeName]
		if !ok {
			val = v.DefaultValue
		}
		str = funcs.ReplaceContent(str, prev, end, v.CodeName, val)
	}

	return str
}

func (this *Js) GetParams() []*JsParam {
	var jps []*JsParam
	db.DB.Model(&JsParam{}).Where("js_id = ?", this.ID).Find(&jps)
	return jps
}

// 根据脚本ID获取脚本
func GetJsById(id int64) *Js {
	if id < 1 {
		return nil
	}
	jsobj := new(Js)
	db.DB.Model(&Js{}).Where("id = ?", id).First(jsobj)
	return jsobj
}

// 根据code获取脚本
func GetJsByCode(code string) *Js {
	if code == "" {
		return nil
	}
	jsobj := new(Js)
	db.DB.Model(&Js{}).Where("code = ?", code).First(jsobj)
	return jsobj
}

func GetJsList(dt *db.ListFinder) ([]*Js, int64) {
	var tks []*Js
	if dt.Page < 1 {
		dt.Page = 1
	}
	if dt.Limit < 1 {
		dt.Limit = 20
	}
	md := db.DB.Model(&Js{})
	if dt.Q != "" {
		qs := fmt.Sprintf("%%%s%%", dt.Q)
		md.Where("title like ?", qs)
	}

	if len(dt.Types) > 0 {
		md.Where("tp in ?", dt.Types)
	}

	if len(dt.Tags) > 0 {
		var taskids []int64
		db.DB.Model(&JsToTag{}).Select("js_id").Where("tag_id in ?", dt.Tags).Find(&taskids)
		if len(taskids) > 0 {
			md.Where("id in ?", taskids)
		}
	}

	var total int64
	md.Count(&total)

	if dt.Scol != "" && dt.By != "" {
		var byy string
		if strings.Contains(dt.By, "desc") {
			byy = "desc"
		} else {
			byy = "asc"
		}
		md.Order(fmt.Sprintf("%s %s", dt.Scol, byy))
	} else {
		md.Order("id DESC")
	}
	md.Offset((dt.Page - 1) * dt.Limit).Limit(dt.Limit).Find(&tks)

	for _, v := range tks {
		// v.Devices = v.GetDevices()
		v.Params = v.GetParams()
		for _, zv := range v.GetTags() {
			v.Tags = append(v.Tags, zv.Name)
		}
	}
	return tks, total
}

func Delete(id any) error {
	return db.DB.Where("id = ?", id).Delete(&Js{}).Error
}

// 通过task添加tags
func (this *Js) AddTags() error {
	tgs := AddTagsBySlice(this.Tags) // 不管三七二十一,将标签在标签表内添加一遍
	var tagIds []int64
	for _, v := range tgs {
		tagIds = append(tagIds, v.ID)
	}

	db.DB.Where("js_id = ?", this.ID).Where("tag_id not in ?", tagIds).Delete(&JsToTag{}) // 不管三七二十一,将对应表中不存在的标签id删除

	if len(tagIds) > 0 {
		tags := make([]*JsToTag, 0, len(tagIds))
		for _, tid := range tagIds {
			tags = append(tags, &JsToTag{
				JsID:  this.ID,
				TagID: tid,
			})
		}

		return db.DB.
			Clauses(clause.OnConflict{
				DoNothing: true,
			}).
			Create(&tags).Error
	}
	return nil
}
