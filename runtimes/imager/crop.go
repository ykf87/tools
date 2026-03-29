package imager

import (
	"fmt"
	"strconv"
)

// 按百分比裁剪,必须是小于1的数
func (c *Crop) output(img *Image) error {
	left := int(float64(img.w) * c.Left)
	top := int(float64(img.h) * c.Top)
	right := int(float64(img.w) * c.Right)
	bottom := int(float64(img.h) * c.Bottom)

	width := img.w - left - right
	height := img.h - top - bottom

	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid crop")
	}
	img.w = width
	img.h = height

	_, err := runVips("crop", img.Src, img.outtemp,
		strconv.Itoa(left),
		strconv.Itoa(top),
		strconv.Itoa(width),
		strconv.Itoa(height),
	)
	return err
}
