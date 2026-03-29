package imager

func (f *Flip) output(img *Image) (err error) {
	hs := "horizontal"
	if f.Type == 2 {
		hs = "vertical"
	}
	_, err = runVips("flip", img.Src, img.outtemp, hs)
	return
}
