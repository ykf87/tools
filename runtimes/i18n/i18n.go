package i18n

import (
	"fmt"
	"strings"
	"sync"
)

var i18nMap sync.Map

func init() {
	fmt.Println("---- i18n 需要同步多语言信息,todo.......")
}

func T(key, lang string) string {
	lang = strings.ToLower(lang)

	v, ok := i18nMap.Load(lang)
	if !ok {
		return key
	}

	val := v.(map[string]string)
	if str, ok := val[key]; ok {
		return str
	}

	return key
}
