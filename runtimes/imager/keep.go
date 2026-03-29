package imager

import (
	"fmt"
)

func (k *KeepWH) output(img *Image) (err error) {
	if img.Width <= 0 && img.Height <= 0 {
		return
	}

	_, err = runVips(
		"thumbnail",
		img.Src, img.outtemp,
		fmt.Sprintf("%d", img.Width),
		"--height", fmt.Sprintf("%d", img.Height),
		"--crop", "centre",
	)

	return
}
