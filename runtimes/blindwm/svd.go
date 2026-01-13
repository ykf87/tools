package blindwm

// 循环嵌入
func embedSVDLoopStrength(mat [][]float64, bits []int, strength float64) {
	h := len(mat)
	w := len(mat[0])
	maxLen := h * w
	for i := 0; i < len(bits); i++ {
		pos := i % maxLen
		x := pos / w
		y := pos % w
		if bits[i] == 1 {
			mat[x][y] += strength
		} else {
			mat[x][y] -= strength
		}
	}
}

// 一次性嵌入小水印
func embedSVDOnceStrength(mat [][]float64, bits []int, strength float64) {
	h := len(mat)
	w := len(mat[0])
	n := len(bits)
	if n > h*w {
		n = h * w
	}
	for i := 0; i < n; i++ {
		x := i / w
		y := i % w
		if bits[i] == 1 {
			mat[x][y] += strength
		} else {
			mat[x][y] -= strength
		}
	}
}

// 提取 SVD bits
func extractSVD(mat [][]float64) []int {
	h := len(mat)
	w := len(mat[0])
	bits := make([]int, h*w)
	for x := 0; x < h; x++ {
		for y := 0; y < w; y++ {
			if mat[x][y] >= 0 {
				bits[x*w+y] = 1
			} else {
				bits[x*w+y] = 0
			}
		}
	}
	return bits
}
