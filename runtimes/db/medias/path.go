package medias

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"tools/runtimes/db"

	"gorm.io/gorm"
)

type MediaPath struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Parent       int64     `json:"parent" gorm:"default:0;index;default:0;uniqueIndex:idx_parent_name"`
	Chain        string    `json:"chain" gorm:"default:'/';index"`                   // 关系链,/间隔
	Name         string    `json:"name" gorm:"not null;uniqueIndex:idx_parent_name"` // 目录名称
	AdminID      int64     `json:"admin_id" gorm:"default:0;index"`                  // 管理员id
	Removed      int       `json:"removed" gorm:"default:0;index"`                   // 是否删除
	Addtime      time.Time `json:"addtime" gorm:"default:0;index"`
	db.BaseModel `json:"-"`
}

var lock sync.Mutex

// 使用路径创建MediaPath表,并返回ID和path
// key 为自己定义的,目录名称对应的唯一值,一般有用户和平台的唯一信息拼接的md5值
func MKDBNameID(fullPath string, adminID int64) (*MediaPath, error) {
	if fullPath == "" {
		return nil, errors.New("path is empty")
	}
	lock.Lock()
	defer lock.Unlock()

	fullPath = strings.Trim(strings.ReplaceAll(fullPath, "\\", "/"), "/")
	re := regexp.MustCompile(`/+`)
	fullPath = re.ReplaceAllString(fullPath, "/")
	fps := strings.Split(fullPath, "/")
	// fmt.Println(fullPath)
	// panic("---")

	var parent int64
	var pids []string
	var lastMph *MediaPath
	for _, v := range fps {
		row := new(MediaPath)
		dbs.DB().Model(&MediaPath{}).Where("name = ? and parent = ?", v, parent).Find(row)

		if row.ID < 1 {
			row = &MediaPath{
				Parent:  parent,
				Name:    v,
				Chain:   re.ReplaceAllString(fmt.Sprintf("/%s/", strings.Join(pids, "/")), "/"),
				AdminID: adminID,
				Addtime: time.Now(),
			}
			err := dbs.Write(func(tx *gorm.DB) error {
				return tx.Create(row).Error
			})
			if err != nil {
				return nil, err
			}
		}
		parent = row.ID
		pids = append(pids, fmt.Sprintf("%d", row.ID))
		lastMph = row
	}
	return lastMph, nil

	// var dbsv []*MediaPath
	// dbs.DB().Model(&MediaPath{}).Where("name in ?", fps).Find(&dbsv)
	// dbmp := make(map[string]*MediaPath)
	// for _, v := range dbsv {
	// 	dbmp[v.Name] = v
	// }

	// var pid int64
	// var pids []string
	// var node *MediaPath

	// var names []string
	// for _, v := range fps {
	// 	if oc, ok := dbmp[v]; ok {
	// 		pid = oc.ID
	// 		pids = append(pids, fmt.Sprintf("%d", oc.ID))
	// 		names = append(names, oc.Name)
	// 		node = oc
	// 	} else {
	// 		obj := &MediaPath{
	// 			Parent:  pid,
	// 			Name:    v,
	// 			Chain:   "/" + strings.Join(pids, "/") + "/",
	// 			Addtime: time.Now(),
	// 		}
	// 		if err := dbs.Write(func(tx *gorm.DB) error {
	// 			return tx.Create(obj).Error
	// 		}); err != nil {
	// 			return nil, err
	// 		} else {
	// 			pid = obj.ID
	// 			pids = append(pids, fmt.Sprintf("%d", obj.ID))
	// 			names = append(names, obj.Name)
	// 		}
	// 		node = obj
	// 	}
	// }
	// return node, nil
}

// 获取当前目录下的子目录
func GetChilds(parent int64, page, limit int, q string) (mps []*MediaPath, total int64) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	md := dbs.DB().Model(&MediaPath{}).Where("parent = ? and removed = 0", parent)

	if q != "" {
		md = md.Where("name like ?", fmt.Sprintf("%%%s%%", q))
	}
	md.Count(&total)

	md = md.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&mps)

	return
}
