package parses

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"tools/runtimes/config"
	"tools/runtimes/funcs"

	"github.com/Rhymond/go-money"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Marshal 处理 struct 和 []struct 的 parse 标签字段
// parse 支持的类型
// datetime;date;time 时间
// number 格式化为显示的数字
// price 格式化为价格格式,自动换算汇率
// medias 格式化图片的完整路径
// email 隐藏邮箱
// phone 隐藏电话
// 如果parse设置了多个,则默认会对更改的字段保留原字段和数据,增加新字段.比如addtime "parse:time;datetime" addtime将保留,而增加addtime_time,addtime_datetime
// -------------------------------------此方法仅支持struct 和 []struct---------------------------
func Marshal(s interface{}, c *gin.Context) (interface{}, error) {
	return formatStructRecursive(reflect.ValueOf(s), c, ""), nil
}

func formatStructRecursive(v reflect.Value, c *gin.Context, parseTag string) interface{} {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return formatStruct(v, c)
	case reflect.Map:
		return formatMap(v, c)
	case reflect.Slice:
		return formatSlice(v, c, parseTag)
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return formatStructRecursive(v.Elem(), c, parseTag)
	default:
		return v.Interface()
	}
}

func formatStruct(v reflect.Value, c *gin.Context) map[string]interface{} {
	t := v.Type()
	result := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		valueField := v.Field(i)

		if !valueField.CanInterface() {
			continue
		}

		tagName := field.Tag.Get("json")
		if tagName == "" || tagName == "-" {
			tagName = field.Name
		}
		parseTag := field.Tag.Get("parse")
		if parseTag == "-" {
			continue
		}

		result[tagName] = handleValue(valueField, c, tagName, parseTag)
	}

	return result
}

func formatMap(v reflect.Value, c *gin.Context) map[string]interface{} {
	result := make(map[string]interface{})
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		result[fmt.Sprint(key.Interface())] = formatStructRecursive(val, c, "")
	}
	return result
}

func formatSlice(v reflect.Value, c *gin.Context, parseTag string) []interface{} {
	var result []interface{}
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		result = append(result, handleValue(elem, c, "", parseTag))
	}
	return result
}

func handleValue(val reflect.Value, c *gin.Context, tagName, parseTag string) interface{} {
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct, reflect.Ptr, reflect.Map, reflect.Slice:
		return formatStructRecursive(val, c, parseTag)
	default:
		rs, ors, err := set(parseTag, val, c)
		if err != nil || parseTag == "" {
			return val.Interface()
		}
		if ors != nil && tagName != "" && parseTag != "" {
			return map[string]interface{}{
				"fmt":    rs,
				"origin": ors,
			}
		}
		return rs
	}
}

func set(str string, val reflect.Value, c *gin.Context) (interface{}, interface{}, error) {
	lang := config.Lang
	timezone := config.Timezone

	switch str {
	case "time", "datetime", "date":
		if val.Kind() == reflect.Int64 {
			if val.Int() < 100 {
				return "", nil, nil
			}
			loc, _ := time.LoadLocation(timezone)
			t := time.Unix(val.Int(), 0).In(loc)
			switch str {
			case "time":
				return t.Format(config.TimeFormat), nil, nil
			case "datetime":
				return t.Format(config.DateTimeFormat), nil, nil
			case "date":
				return t.Format(config.DateFormat), nil, nil
			}
		}
		return val.Interface(), nil, nil
	case "number":
		if val.Kind() == reflect.Int64 {
			l, _ := language.Parse(lang)
			p := message.NewPrinter(l)
			return p.Sprintf("%d", val.Int()), val.Int(), nil
		}
		return val.Interface(), nil, nil
	case "price":
		if val.Kind() == reflect.Float64 {
			final := funcs.TruncateFloat64(val.Float()*config.CurrRate, 2)
			return money.NewFromFloat(final, config.Currency).Display(), final, nil
		}
		return val.Interface(), nil, nil
	case "medias":
		if val.Kind() == reflect.String {
			ss := strings.Split(val.String(), ",")
			for i, v := range ss {
				ss[i] = fmt.Sprintf("%s/%s", config.MediaUrl, v)
			}
			return strings.Join(ss, ","), nil, nil
		}
		return val.Interface(), nil, nil
	case "json":
		if val.Kind() == reflect.String {
			var mp map[string]interface{}
			if err := json.Unmarshal([]byte(val.String()), &mp); err != nil {
				return nil, nil, err
			}
			return mp, nil, nil
		}
		return val.Interface(), nil, nil
	case "phone":
		if val.Kind() == reflect.String {
			return maskPhone(val.String()), nil, nil
		}
		return val.Interface(), nil, nil
	case "email":
		if val.Kind() == reflect.String {
			return maskEmail(val.String()), nil, nil
		}
		return val.Interface(), nil, nil
	}
	return val.Interface(), nil, nil
}

func maskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 1 {
		return email
	}
	name := email[:at]
	domain := email[at:]
	if len(name) <= 3 {
		return name[:1] + "***" + domain
	}
	return name[:1] + strings.Repeat("*", len(name)-2) + name[len(name)-1:] + domain
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + strings.Repeat("*", len(phone)-7) + phone[len(phone)-4:]
}
