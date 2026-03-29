package imager

import "strconv"

func (g *Gamma) output(img *Image) (err error) {
	_, err = runVips("gamma", img.Src, img.outtemp, "--exponent", strconv.FormatFloat(g.Value, 'f', -1, 64))
	return
}
