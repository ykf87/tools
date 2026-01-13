package blindwm

func DWT2D(mat [][]float64) (LL, LH, HL, HH [][]float64) {
	h := len(mat)
	w := len(mat[0])

	if h%2 != 0 {
		h++
	}
	if w%2 != 0 {
		w++
	}

	LL = makeMatrix(h/2, w/2)
	LH = makeMatrix(h/2, w/2)
	HL = makeMatrix(h/2, w/2)
	HH = makeMatrix(h/2, w/2)

	for y := 0; y < h; y += 2 {
		for x := 0; x < w; x += 2 {
			a := getSafe(mat, y, x)
			b := getSafe(mat, y, x+1)
			c := getSafe(mat, y+1, x)
			d := getSafe(mat, y+1, x+1)

			LL[y/2][x/2] = (a + b + c + d) / 2
			LH[y/2][x/2] = (a - b + c - d) / 2
			HL[y/2][x/2] = (a + b - c - d) / 2
			HH[y/2][x/2] = (a - b - c + d) / 2
		}
	}
	return
}

func IDWT2D(LL, LH, HL, HH [][]float64) [][]float64 {
	h := len(LL) * 2
	w := len(LL[0]) * 2
	mat := makeMatrix(h, w)

	for y := 0; y < len(LL); y++ {
		for x := 0; x < len(LL[0]); x++ {
			a := (LL[y][x] + LH[y][x] + HL[y][x] + HH[y][x]) / 2
			b := (LL[y][x] - LH[y][x] + HL[y][x] - HH[y][x]) / 2
			c := (LL[y][x] + LH[y][x] - HL[y][x] - HH[y][x]) / 2
			d := (LL[y][x] - LH[y][x] - HL[y][x] + HH[y][x]) / 2

			mat[2*y][2*x] = a
			mat[2*y][2*x+1] = b
			mat[2*y+1][2*x] = c
			mat[2*y+1][2*x+1] = d
		}
	}
	return mat
}

func getSafe(mat [][]float64, y, x int) float64 {
	h := len(mat)
	w := len(mat[0])
	if y >= h {
		y = h - 1
	}
	if x >= w {
		x = w - 1
	}
	return mat[y][x]
}
