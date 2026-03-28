package imager

import (
	"fmt"
	"strconv"
)

func (c *Crop) output(input, output string) error {
	meta, _ := vipsheader(input) // 你自己实现获取宽高
	w, h := meta.Width, meta.Height

	left := int(float64(w) * c.Left)
	top := int(float64(h) * c.Top)
	right := int(float64(w) * c.Right)
	bottom := int(float64(h) * c.Bottom)

	width := w - left - right
	height := h - top - bottom

	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid crop")
	}

	_, err := runVips("crop", input, output,
		strconv.Itoa(left),
		strconv.Itoa(top),
		strconv.Itoa(width),
		strconv.Itoa(height),
	)
	return err
}
