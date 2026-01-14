package wmimage

import (
	"image"
	"image/color"
	"log"

	"github.com/disintegration/imaging"
)

func DetectFromFile(path string, seed int64) float64 {
	// 打开图片
	img, err := imaging.Open(path)
	if err != nil {
		log.Printf("打开图片失败: %v", err)
		return 0
	}

	// 转成 YCbCr
	bounds := img.Bounds()
	sizeX := bounds.Dx()
	// sizeY := bounds.Dy()
	ycbcr := image.NewYCbCr(bounds, image.YCbCrSubsampleRatio444)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rr := uint8(r >> 8)
			gg := uint8(g >> 8)
			bb := uint8(b >> 8)
			Y, Cb, Cr := color.RGBToYCbCr(rr, gg, bb)
			ycbcr.Y[y*sizeX+x] = Y
			ycbcr.Cb[y*sizeX+x] = Cb
			ycbcr.Cr[y*sizeX+x] = Cr
		}
	}

	// 调用原 Detect 方法
	return Detect(ycbcr, seed)
}

func Detect(img *image.YCbCr, seed int64) float64 {
	pnSeq := pn(seed, len(midFreq))
	var score float64
	count := 0
	size := img.Rect.Dx()

	for by := 0; by < size; by += 8 {
		for bx := 0; bx < size; bx += 8 {
			var block [8][8]float64
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					block[y][x] = float64(img.Y[(by+y)*img.YStride+(bx+x)])
				}
			}

			d := dct8(block)
			var sum float64
			for i, p := range midFreq {
				sum += d[p[0]][p[1]] * pnSeq[i]
			}
			score += sum
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return score / float64(count)
}
