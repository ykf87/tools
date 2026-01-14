package wmimage

func pn(seed int64, n int) []float64 {
	r := seed
	out := make([]float64, n)
	for i := range out {
		r = (r*1103515245 + 12345) & 0x7fffffff
		if r&1 == 0 {
			out[i] = 1
		} else {
			out[i] = -1
		}
	}
	return out
}
