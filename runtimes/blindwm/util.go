package blindwm

import (
	"image"
	"image/color"
)

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func majority(a, b, c int) int {
	if a+b+c >= 2 {
		return 1
	}
	return 0
}

func imageToMatrix(img image.Image) (r, g, b [][]float64) {
	bounds := img.Bounds()
	h := bounds.Dy()
	w := bounds.Dx()
	r = makeMatrix(h, w)
	g = makeMatrix(h, w)
	b = makeMatrix(h, w)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rr, gg, bb, _ := img.At(x, y).RGBA()
			r[y][x] = float64(rr >> 8)
			g[y][x] = float64(gg >> 8)
			b[y][x] = float64(bb >> 8)
		}
	}
	return
}

func matrixToImage(r, g, b [][]float64) image.Image {
	h := len(r)
	w := len(r[0])
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			out.SetRGBA(x, y, color.RGBA{
				R: clamp(r[y][x]),
				G: clamp(g[y][x]),
				B: clamp(b[y][x]),
				A: 255,
			})
		}
	}
	return out
}

func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func makeMatrix(h, w int) [][]float64 {
	mat := make([][]float64, h)
	for i := range mat {
		mat[i] = make([]float64, w)
	}
	return mat
}
