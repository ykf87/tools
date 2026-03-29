package imager

import "strconv"

func (r *Resize) output(img *Image) (err error) {
	if r.Scale <= 0 {
		r.Scale = 1
	}
	_, err = runVips("resize", img.Src, img.outtemp, strconv.FormatFloat(r.Scale, 'f', -1, 64))
	return
}
