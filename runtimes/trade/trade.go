package trade

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var platforms sync.Map
var loaded bool

func Init() (mp map[string]*Api, err error) {
	if loaded == true {
		mp = configMap()
	} else {
		mp, err = getConfig()
		if err == nil {
			loaded = true
		}
	}

	return
}

func fileName(dir string) string {
	if dir != "" {
		if _, err := os.Stat(dir); err != nil {
			os.MkdirAll(dir, os.ModePerm)
		}
	}

	return filepath.Join(dir, ".trade")
}

func configMap() map[string]*Api {
	mp := make(map[string]*Api)
	platforms.Range(func(key, value any) bool {
		k, kok := key.(string)
		v, vok := value.(*Api)

		if kok && vok && k != "" {
			mp[k] = v
		}

		return true
	})
	return mp
}

// 添加一行数据
func AddRow(k string, val *Api) {
	platforms.Store(k, val)
}

// 获取配置
func getConfig() (map[string]*Api, error) {
	fn := fileName("")

	if _, err := os.Stat(fn); err != nil {
		return nil, err
	}

	bt, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	str := AesDecryptCBC(string(bt))
	mp := make(map[string]*Api)
	if str != "" {
		if err := json.Unmarshal([]byte(str), &mp); err != nil {
			return nil, err
		}
	}

	for k, v := range mp {
		platforms.Store(k, v)
	}

	return mp, nil
}

// 保存配置
func SaveConfig() error {
	data, err := json.Marshal(configMap())
	if err != nil {
		return err
	}

	fc := AesEncryptCBC(string(data))

	return os.WriteFile(fileName(""), []byte(fc), os.ModePerm)
}
