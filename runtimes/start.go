// 启动的必须检查
package runtimes

import (
	"os"
	"tools/runtimes/config"
	"tools/runtimes/funcs"
)

func init() {
	config.RuningRoot = funcs.RunnerPath()

	for _, v := range config.Mkdirs {
		full := config.FullPath(v.DirName)
		if _, err := os.Stat(full); err != nil {
			if err := os.MkdirAll(full, v.Mode); err != nil {
				panic(err)
			}
			if v.IsHide == true {
				funcs.HiddenDir(full)
			}
		}
	}
}
