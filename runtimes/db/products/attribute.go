package products

import (
	"fmt"
	"strings"
	"tools/runtimes/db"

	"gorm.io/gorm"
)

// 属性表
type Attribute struct {
	ID     int64 `gorm:"primaryKey"`
	Sort   int   `gorm:"default:0;index"`
	Status int8  `gorm:"default:1;index"` // 1启用 0禁用
	Type   int8  `gorm:"default:0"`       // 1 = 展示属性,2 = SKU属性,3 = 定制属性

	// 多语言
	Langs []AttributeLang `gorm:"foreignKey:AttributeID;constraint:OnDelete:CASCADE"`

	// 属性值
	Values []AttributeValue `gorm:"foreignKey:AttributeID;constraint:OnDelete:CASCADE"`
}

type AttributeLang struct {
	AttributeID int64  `gorm:"primaryKey;not null"`
	Lang        string `gorm:"primaryKey;size:10"`
	Name        string `gorm:"size:100;index"`
}

type AttributeValue struct {
	ID          int64 `gorm:"primaryKey"`
	AttributeID int64 `gorm:"not null;index"`

	Sort   int  `gorm:"default:0;index"`
	Status int8 `gorm:"default:1;index"`

	Langs []AttributeValueLang `gorm:"foreignKey:AttributeValueID;constraint:OnDelete:CASCADE"`
}
type AttributeValueLang struct {
	AttributeValueID int64  `gorm:"primaryKey;not null"`
	Lang             string `gorm:"primaryKey;size:10"`
	Name             string `gorm:"size:100;index"`
}

// 产品属性表
type ProductAttribute struct {
	ID          int64 `gorm:"primaryKey"`
	ProductID   int64 `gorm:"uniqueIndex:idx_product_attr"`
	AttributeID int64 `gorm:"uniqueIndex:idx_product_attr"`
	// 属性信息（方便 preload）
	Attribute Attribute `gorm:"foreignKey:AttributeID"`
	// 选中的值
	Values []ProductAttributeValue `gorm:"foreignKey:ProductAttributeID;constraint:OnDelete:CASCADE"`
}
type ProductAttributeValue struct {
	ID                 int64 `gorm:"primaryKey"`
	ProductAttributeID int64 `gorm:"index"`
	AttributeValueID   int64 `gorm:"index"`
	// 可选：支持自定义值（比如手填）
	CustomValue string `gorm:"size:255"`
	// 方便 preload
	Value AttributeValue `gorm:"foreignKey:AttributeValueID"`
}

// 筛选加速表
type ProductAttrIndex struct {
	ProductID   int64 `gorm:"index"`
	AttributeID int64 `gorm:"index"`
	ValueID     int64 `gorm:"index"`
}

// 查询属性
type AttributeDTO struct {
	ID     int64             `json:"id"`
	Name   string            `json:"name"`
	Values []AttributeValDTO `json:"values,omitempty"`
}

type AttributeValDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func getLangName(langs []AttributeLang, lang string) string {
	var fallback string

	for _, l := range langs {
		if l.Lang == lang {
			return l.Name
		}
		if l.Lang == "en" {
			fallback = l.Name
		}
	}

	if fallback != "" {
		return fallback
	}

	// 最后兜底：随便拿一个
	if len(langs) > 0 {
		return langs[0].Name
	}

	return ""
}

func getValueLangName(langs []AttributeValueLang, lang string) string {
	var fallback string

	for _, l := range langs {
		if l.Lang == lang {
			return l.Name
		}
		if l.Lang == "en" {
			fallback = l.Name
		}
	}

	if fallback != "" {
		return fallback
	}

	if len(langs) > 0 {
		return langs[0].Name
	}

	return ""
}

func GetAttributeList(q db.ListFinder) ([]Attribute, error) {
	var attrs []Attribute

	query := DB.DB().Model(&Attribute{}).
		Where("status = 1")

	// 类型筛选
	if q.Type > 0 {
		query = query.Where("type = ?", q.Type)
	}

	// 搜索（多语言）
	if q.Q != "" {
		query = query.Joins("JOIN attribute_langs al ON al.attribute_id = attributes.id").
			Where("al.name ILIKE ?", "%"+q.Q+"%")
	}

	// === 语言处理 ===
	if q.Lang != "" {
		// 单语言 + fallback（en）
		query = query.Preload("Langs", "lang IN ?", []string{q.Lang, "en"})
	} else {
		// 全部语言
		query = query.Preload("Langs")
	}

	// === Values ===
	if q.WithValues {
		query = query.Preload("Values", func(tx *gorm.DB) *gorm.DB {
			return tx.
				Where("status = 1").
				Order("id DESC")
		})

		if q.Lang != "" {
			query = query.Preload("Values.Langs", "lang IN ?", []string{q.Lang, "en"})
		} else {
			query = query.Preload("Values.Langs")
		}
	}

	err := query.Order("sort ASC").Order("id DESC").Find(&attrs).Error
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

func BuildAttributeDTO(attrs []Attribute, lang string) []AttributeDTO {
	var result []AttributeDTO

	for _, attr := range attrs {
		dto := AttributeDTO{
			ID:   attr.ID,
			Name: getLangName(attr.Langs, lang),
		}

		for _, v := range attr.Values {
			val := AttributeValDTO{
				ID:   v.ID,
				Name: getValueLangName(v.Langs, lang),
			}
			dto.Values = append(dto.Values, val)
		}

		result = append(result, dto)
	}

	return result
}

func BuildAttributeMap(attrs []Attribute) map[string][]AttributeDTO {
	result := make(map[string][]AttributeDTO)

	for _, attr := range attrs {

		// 找这个属性支持的语言
		for _, langItem := range attr.Langs {

			dto := AttributeDTO{
				ID:   attr.ID,
				Name: langItem.Name,
			}

			// values
			for _, v := range attr.Values {
				for _, vl := range v.Langs {
					if vl.Lang == langItem.Lang {
						dto.Values = append(dto.Values, AttributeValDTO{
							ID:   v.ID,
							Name: vl.Name,
						})
					}
				}
			}

			result[langItem.Lang] = append(result[langItem.Lang], dto)
		}
	}

	return result
}

// 新增或修改
type MultiLangAttr struct {
	ID int64

	Langs map[string]string

	Values []struct {
		ID    int64
		Langs map[string]string
	}
}

func UpsertAttributes(inputs []MultiLangAttr) error {
	// if err := validateAttrs(inputs); err != nil {
	// 	return err
	// }
	return DB.DB().Transaction(func(tx *gorm.DB) error {

		for _, item := range inputs {
			if len(item.Langs) == 0 {
				continue
			}
			// =========================
			// 1️⃣ 处理 Attribute
			// =========================
			var attr Attribute
			var err error

			if item.ID > 0 {
				err = tx.First(&attr, item.ID).Error
				if err != nil {
					return err
				}
			} else {
				// 尝试用任意语言查（仅限当前 Attribute）
				found := false
				for lang, name := range item.Langs {
					if lang == "" || name == "" {
						continue
					}

					var langRow AttributeLang
					err := tx.Where("lang = ? AND name = ?", lang, name).
						First(&langRow).Error

					if err == nil {
						if err := tx.First(&attr, langRow.AttributeID).Error; err == nil {
							found = true
							return fmt.Errorf("属性 %s 已存在,请在所在行添加值", name)
						}
					}
				}

				if !found {
					attr = Attribute{}
					if err := tx.Create(&attr).Error; err != nil {
						return err
					}
				}
			}

			// =========================
			// 2️⃣ 多语言（Attribute）
			// =========================
			for lang, name := range item.Langs {
				if name == "" {
					continue
				}

				var langRow AttributeLang
				err := tx.Where("attribute_id = ? AND lang = ?", attr.ID, lang).
					First(&langRow).Error

				if err != nil {
					if err := tx.Create(&AttributeLang{
						AttributeID: attr.ID,
						Lang:        lang,
						Name:        name,
					}).Error; err != nil {
						return err
					}
				} else {
					if err := tx.Model(&langRow).Update("name", name).Error; err != nil {
						return err
					}
				}
			}

			// =========================
			// 3️⃣ 处理 Attribute Values
			// =========================
			for _, v := range item.Values {
				if len(v.Langs) < 1 {
					continue
				}
				var val AttributeValue
				valFound := false

				// 👉 优先用 ID
				if v.ID > 0 {
					err := tx.First(&val, v.ID).Error
					if err == nil {
						// ⚠️ 校验归属
						if val.AttributeID != attr.ID {
							return fmt.Errorf("value %d 不属于 attribute %d", val.ID, attr.ID)
						}
						valFound = true
					}
				}

				// 👉 没 ID，用当前 Attribute + name 查
				if !valFound {
					for lang, name := range v.Langs {
						if name == "" || lang == "" {
							continue
						}

						err := tx.Joins("JOIN attribute_value_langs avl ON avl.attribute_value_id = attribute_values.id").
							Where("attribute_values.attribute_id = ?", attr.ID).
							Where("avl.lang = ? AND avl.name = ?", lang, name).
							First(&val).Error

						if err == nil {
							valFound = true
							break
						}
					}
				}

				// 👉 不存在就创建
				if !valFound {
					val = AttributeValue{
						AttributeID: attr.ID,
					}
					if err := tx.Create(&val).Error; err != nil {
						return err
					}
				}

				// =========================
				// 4️⃣ 多语言（Value）
				// =========================
				for lang, name := range v.Langs {
					if name == "" || lang == "" {
						continue
					}

					var valLang AttributeValueLang
					err := tx.Where("attribute_value_id = ? AND lang = ?", val.ID, lang).
						First(&valLang).Error

					if err != nil {
						if err := tx.Create(&AttributeValueLang{
							AttributeValueID: val.ID,
							Lang:             lang,
							Name:             name,
						}).Error; err != nil {
							return err
						}
					} else {
						if err := tx.Model(&valLang).Update("name", name).Error; err != nil {
							return err
						}
					}
				}
			}
		}

		return nil
	})
}

func validateAttrs(inputs []MultiLangAttr) error {
	for i, item := range inputs {

		// 至少要有一个语言
		if len(item.Langs) == 0 {
			return fmt.Errorf("attr[%d] langs 不能为空", i)
		}

		validLang := false
		for lang, name := range item.Langs {
			if strings.TrimSpace(lang) != "" && strings.TrimSpace(name) != "" {
				validLang = true
				break
			}
		}

		if !validLang {
			return fmt.Errorf("attr[%d] 至少需要一个有效语言", i)
		}

		// values 校验
		for j, v := range item.Values {
			if len(v.Langs) == 0 {
				return fmt.Errorf("attr[%d].value[%d] langs 不能为空", i, j)
			}

			valid := false
			for lang, name := range v.Langs {
				if strings.TrimSpace(lang) != "" && strings.TrimSpace(name) != "" {
					valid = true
					break
				}
			}

			if !valid {
				return fmt.Errorf("attr[%d].value[%d] 至少需要一个有效语言", i, j)
			}
		}
	}
	return nil
}
