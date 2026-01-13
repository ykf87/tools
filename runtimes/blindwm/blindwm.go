package blindwm

import (
	"errors"
	"image"
	"image/color"
)

var ErrExtractFailed = errors.New("extract failed")

// EmbedImage 嵌入水印，payload 最多不超过图片像素数*3/8 字节
func EmbedImage(src image.Image, payload []byte) (image.Image, error) {
	if len(payload) == 0 {
		return src, nil
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	maxBits := w * h * 3
	if len(payload)*8*3 > maxBits {
		return nil, errors.New("payload too large for image")
	}

	// 前 32 bit 长度 + payload bits
	lenBits := Uint32ToBits(uint32(len(payload)))
	payloadBits := BytesToBits(payload)
	allBits := append(lenBits, payloadBits...)
	allBits = RepeatBits(allBits, 3) // 简单 ECC，重复 3 次

	out := image.NewRGBA(bounds)

	bitIdx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r0, g0, b0, a := src.At(x, y).RGBA()
			r := uint8(r0 >> 8)
			g := uint8(g0 >> 8)
			b := uint8(b0 >> 8)

			// 嵌入 RGB 最低位
			if bitIdx < len(allBits) {
				r = (r & 0xFE) | uint8(allBits[bitIdx])
				bitIdx++
			}
			if bitIdx < len(allBits) {
				g = (g & 0xFE) | uint8(allBits[bitIdx])
				bitIdx++
			}
			if bitIdx < len(allBits) {
				b = (b & 0xFE) | uint8(allBits[bitIdx])
				bitIdx++
			}
			out.Set(x, y, color.RGBA{R: r, G: g, B: b, A: uint8(a >> 8)})
		}
	}

	return out, nil
}

// ExtractImage 提取水印
func ExtractImage(src image.Image) ([]byte, error) {
	bounds := src.Bounds()
	// w, h := bounds.Dx(), bounds.Dy()
	allBits := []int{}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r0, g0, b0, _ := src.At(x, y).RGBA()
			r := int((r0 >> 8) & 1)
			g := int((g0 >> 8) & 1)
			b := int((b0 >> 8) & 1)
			allBits = append(allBits, r, g, b)
		}
	}

	allBits = DecodeRepeatedBits(allBits, 3) // 解码重复码

	if len(allBits) < 32 {
		return nil, ErrExtractFailed
	}

	length := BitsToUint32(allBits[:32])
	available := (len(allBits) - 32) / 8
	if int(length) > available {
		length = uint32(available) // 容错
	}

	data := BitsToBytesLen(allBits[32:], int(length))
	return data, nil
}
