package audios

import (
	"fmt"
	"tools/runtimes/db"

	"gorm.io/gorm"
)

func GetAudioTags(s *db.ListFinder) (total int64, lists []*AudioTag) {

	model := Dbs.DB().Model(&AudioTag{})
	if s.Q != "" {
		model = model.Where("name like ?", fmt.Sprintf("%%%s%%", s.Q))
	}
	model.Count(&total)

	if s.Limit > 0 {
		page := 1
		if s.Page > 0 {
			page = s.Page
		}
		model = model.Offset((page - 1) * s.Limit).Limit(s.Limit)
	}

	model.Order("id DESC").Find(&lists)
	return
}

func MakerTags(names []string) ([]*AudioTag, error) {
	if len(names) == 0 {
		return []*AudioTag{}, nil
	}

	// 1️⃣ 去重（很关键）
	nameSet := make(map[string]struct{}, len(names))
	uniqueNames := make([]string, 0, len(names))
	for _, n := range names {
		if n == "" {
			continue
		}
		if _, ok := nameSet[n]; !ok {
			nameSet[n] = struct{}{}
			uniqueNames = append(uniqueNames, n)
		}
	}

	// 2️⃣ 查已有的
	var existing []*AudioTag
	if err := Dbs.DB().Where("name IN ?", uniqueNames).Find(&existing).Error; err != nil {
		return nil, err
	}

	// 3️⃣ 找缺失的
	existMap := make(map[string]*AudioTag, len(existing))
	for _, t := range existing {
		existMap[t.Name] = t
	}

	toCreate := make([]*AudioTag, 0)
	for _, name := range uniqueNames {
		if _, ok := existMap[name]; !ok {
			toCreate = append(toCreate, &AudioTag{Name: name})
		}
	}

	// 4️⃣ 批量创建（关键：避免一条条 insert）
	if len(toCreate) > 0 {
		if err := Dbs.Write(func(tx *gorm.DB) error {
			return tx.Create(&toCreate).Error
		}); err != nil {
			return nil, err
		}
	}

	// 5️⃣ 合并结果（保证顺序可控）
	result := make([]*AudioTag, 0, len(uniqueNames))
	for _, name := range uniqueNames {
		if t, ok := existMap[name]; ok {
			result = append(result, t)
		} else {
			// 新创建的
			for _, t2 := range toCreate {
				if t2.Name == name {
					result = append(result, t2)
					break
				}
			}
		}
	}

	return result, nil
}
