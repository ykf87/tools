package imager

import (
	"strconv"
	"strings"
)

func (l *Linear) output(i, o string) (err error) {
	tmp := strings.ReplaceAll(i, ".", "linear.")
	if l.Brightness == 0 {
		l.Brightness = 1.0
	}
	if l.Contrast == 0 {
		l.Contrast = 1.0
	}
	_, err = runVips("linear", i, tmp, strconv.FormatFloat(l.Brightness, 'f', -1, 64), strconv.FormatFloat(l.Contrast, 'f', -1, 64))
	return
}
