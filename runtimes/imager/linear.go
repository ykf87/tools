package imager

import (
	"strconv"
)

func (l *Linear) output(img *Image) (err error) {
	if l.Brightness < -100 {
		l.Brightness = -100
	} else if l.Brightness > 100 {
		l.Brightness = 100
	}
	if l.Contrast < 0.5 {
		l.Contrast = 1
	} else if l.Contrast > 2 {
		l.Contrast = 2
	}

	bt := strconv.FormatFloat(l.Contrast, 'f', -1, 64)
	ct := strconv.FormatFloat(l.Brightness, 'f', -1, 64)

	_, err = runVips("linear", img.Src, img.outtemp, "--", bt, ct)
	return
}
