package clearer

import (
	"tools/runtimes/funcs"
)

// 让图片更清晰
func Clearers(src, output, modules string) (str string, err error) {
	if modules == "" {
		modules = DEFMODEL
	}
	// fmt.Println(str, err, "变清晰...")
	str, _, err = funcs.RunCommand(true, FullFileName,
		"-i", src,
		"-o", output,
		"-n", modules,
	)
	return
}
