package wmimage

import "math"

func dct8(block [8][8]float64) [8][8]float64 {
	var out [8][8]float64
	for u := 0; u < 8; u++ {
		for v := 0; v < 8; v++ {
			sum := 0.0
			for x := 0; x < 8; x++ {
				for y := 0; y < 8; y++ {
					sum += block[x][y] *
						math.Cos((2*float64(x)+1)*float64(u)*math.Pi/16) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/16)
				}
			}
			cu, cv := 1.0, 1.0
			if u == 0 {
				cu = 1 / math.Sqrt2
			}
			if v == 0 {
				cv = 1 / math.Sqrt2
			}
			out[u][v] = 0.25 * cu * cv * sum
		}
	}
	return out
}

func idct8(block [8][8]float64) [8][8]float64 {
	var out [8][8]float64
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			sum := 0.0
			for u := 0; u < 8; u++ {
				for v := 0; v < 8; v++ {
					cu, cv := 1.0, 1.0
					if u == 0 {
						cu = 1 / math.Sqrt2
					}
					if v == 0 {
						cv = 1 / math.Sqrt2
					}
					sum += cu * cv * block[u][v] *
						math.Cos((2*float64(x)+1)*float64(u)*math.Pi/16) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/16)
				}
			}
			out[x][y] = 0.25 * sum
		}
	}
	return out
}
