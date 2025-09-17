package i18n

import (
	"fmt"
	"sync"
	"tools/runtimes/config"
)

var i18nMap sync.Map

func init() {
	fmt.Println("---- i18n 需要同步多语言信息,todo.......")
}

func T(key string, args ...any) string {
	v, ok := i18nMap.Load(config.Lang)
	if !ok {
		return fmt.Sprintf(key, args...)
	}

	val := v.(map[string]string)
	if str, ok := val[key]; ok {
		return fmt.Sprintf(str, args...)
	}

	return fmt.Sprintf(key, args...)
}
