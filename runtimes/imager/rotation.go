package imager

import "strconv"

func (r *Rotation) output(img *Image) (err error) {
	_, err = runVips("rotate", img.Src, img.outtemp, "--", strconv.FormatFloat(r.Angle, 'f', -1, 64))
	return
}
