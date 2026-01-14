package wmimage

func haarDWT2D(block [][]float64) (LL, LH, HL, HH [][]float64) {
	n := len(block)
	LL = make([][]float64, n/2)
	LH = make([][]float64, n/2)
	HL = make([][]float64, n/2)
	HH = make([][]float64, n/2)
	for i := range LL {
		LL[i] = make([]float64, n/2)
		LH[i] = make([]float64, n/2)
		HL[i] = make([]float64, n/2)
		HH[i] = make([]float64, n/2)
	}
	for i := 0; i < n; i += 2 {
		for j := 0; j < n; j += 2 {
			a := block[i][j]
			b := block[i][j+1]
			c := block[i+1][j]
			d := block[i+1][j+1]
			LL[i/2][j/2] = (a + b + c + d) / 2
			LH[i/2][j/2] = (a - b + c - d) / 2
			HL[i/2][j/2] = (a + b - c - d) / 2
			HH[i/2][j/2] = (a - b - c + d) / 2
		}
	}
	return
}

// 逆 Haar DWT
func ihaarDWT2D(LL, LH, HL, HH [][]float64) [][]float64 {
	n := len(LL) * 2
	out := make([][]float64, n)
	for i := range out {
		out[i] = make([]float64, n)
	}
	for i := 0; i < len(LL); i++ {
		for j := 0; j < len(LL); j++ {
			a := (LL[i][j] + LH[i][j] + HL[i][j] + HH[i][j]) / 2
			b := (LL[i][j] - LH[i][j] + HL[i][j] - HH[i][j]) / 2
			c := (LL[i][j] + LH[i][j] - HL[i][j] - HH[i][j]) / 2
			d := (LL[i][j] - LH[i][j] - HL[i][j] + HH[i][j]) / 2
			out[i*2][j*2] = a
			out[i*2][j*2+1] = b
			out[i*2+1][j*2] = c
			out[i*2+1][j*2+1] = d
		}
	}
	return out
}
