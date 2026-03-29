package imager

import "strconv"

func (g *Gaussblur) output(img *Image) (err error) {
	if g.Value < 0 {
		g.Value = 0
	}
	_, err = runVips("gaussblur", img.Src, img.outtemp, "--", strconv.FormatFloat(g.Value, 'f', -1, 64))
	return
}
