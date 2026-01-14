package wmimage

import (
	"image"
	"image/color"

	"github.com/disintegration/imaging"
)

func preprocessYCbCr(img image.Image, size int) *image.YCbCr {
	resized := imaging.Resize(img, size, size, imaging.Lanczos)
	bounds := resized.Bounds()
	ycbcr := image.NewYCbCr(bounds, image.YCbCrSubsampleRatio444)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			rr := uint8(r >> 8)
			gg := uint8(g >> 8)
			bb := uint8(b >> 8)
			y1, cb, cr := color.RGBToYCbCr(rr, gg, bb)
			ycbcr.Y[y*size+x] = y1
			ycbcr.Cb[y*size+x] = cb
			ycbcr.Cr[y*size+x] = cr
		}
	}
	return ycbcr
}
