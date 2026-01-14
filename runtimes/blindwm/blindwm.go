package blindwm

import (
	"github.com/disintegration/imaging"
	"github.com/kirklin/go-blind-watermark/bwm"
)

func AddImageTxt(input, output, text string) (int, error) {
	src, err := imaging.Open(input)
	if err != nil {
		return 0, err
	}
	wmBits := bwm.TextToBits(text)
	engine := bwm.New(12345, 67890)
	engine.D1 = 40.0
	watermarkedImg, err := engine.Embed(src, wmBits)
	if err != nil {
		return 0, err
	}
	imaging.Save(watermarkedImg, output)
	return len(wmBits), nil
}

func DecodeImageTxt(imgSrc string, length int) (string, error) {
	src, err := imaging.Open(imgSrc)
	if err != nil {
		return "", err
	}

	engine := bwm.New(12345, 67890)
	// 设置强度 (D1越大越抗攻击，但画质越差)
	engine.D1 = 40.0

	extractedBits, _ := engine.Extract(src, length)
	return bwm.BitsToText(extractedBits), nil
}
