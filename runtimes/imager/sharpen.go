package imager

// 锐化功能在cli不能用
func (s *Sharpen) output(img *Image) (err error) {
	if s.Value < 0 {
		s.Value = 1
	}
	_, err = runVips("sharpen", img.Src, img.outtemp)
	return
}
