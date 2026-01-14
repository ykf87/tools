package wmimage

import (
	"image"
)

var midFreq = [][2]int{{3, 4}, {4, 3}, {4, 4}, {5, 3}, {3, 5}}

func Embed(src image.Image, cfg Config) *image.YCbCr {
	img := preprocessYCbCr(src, cfg.TargetSize)
	pnSeq := pn(cfg.Seed, len(midFreq))
	size := cfg.TargetSize

	// RS 编码占位
	_ = newRS(cfg.RSData, cfg.RSPare)

	// 随机块顺序
	blockOrder := randomBlockOrder(size/8*size/8, cfg.Seed)

	for _, idx := range blockOrder {
		bx := (idx % (size / 8)) * 8
		by := (idx / (size / 8)) * 8

		var block [8][8]float64
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				block[y][x] = float64(img.Y[(by+y)*img.YStride+(bx+x)])
			}
		}

		if blockEnergy(block) < cfg.MinEnergy {
			continue
		}

		d := dct8(block)
		scale := cfg.BaseStrength * (blockEnergy(block) / 2000.0)

		for i, p := range midFreq {
			d[p[0]][p[1]] += scale * pnSeq[i]
		}

		out := idct8(d)
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				v := out[y][x]
				if v < 0 {
					v = 0
				}
				if v > 255 {
					v = 255
				}
				img.Y[(by+y)*img.YStride+(bx+x)] = uint8(v)
			}
		}
	}
	return img
}
