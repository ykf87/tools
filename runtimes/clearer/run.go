package clearer

import (
	"log"
	"tools/runtimes/funcs"
)

// 让图片更清晰
func Clearers(src, output, modules string) (str string, err error) {
	if modules == "" {
		modules = DEFMODEL
	}
	log.Println("开始执行")
	str, _, err = funcs.RunCommand(true, FullFileName,
		"-i", src,
		"-o", output,
		"-n", modules,
	)
	log.Println("执行完成")
	return
}
