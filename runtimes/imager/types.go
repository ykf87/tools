package imager

import (
	"fmt"
	"regexp"
	"strconv"
	"tools/runtimes/funcs"
	"tools/runtimes/libvips"
)

// 裁剪,上右下左各裁了多少,最终应该是按百分比计算
type Crop struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// 翻转,水平和垂直
type Flip struct {
	Type int `json:"type"` // 1水平翻转,2垂直翻转
}

// 仿射变换
type Affine struct {
}

// 透视变换
type Mapim struct {
}

// 缩放
type Resize struct {
	Scale float64 `json:"scale"` // >0
}

// 旋转,角度
type Rotation struct {
	Angle float64 `json:"angle"`
}

type Linear struct {
	Brightness float64 `json:"brightness"` // 调节亮度 -100 ~ 100
	Contrast   float64 `json:"contrast"`   // 调节对比度 0.5 ~ 2
}

// // 调节亮度
// type Brightness struct {
// 	Value float64 `json:"value"` // -100 ~ 100
// }

// // 调节对比度
// type Contrast struct {
// 	Value float64 `json:"value"` // 0.5 ~ 2
// }

// 调节饱和度
type Saturation struct {
	Value float64 `json:"value"` // 0 ~ 3
}

// 调节gamma线
type Gamma struct {
	Value float64 `json:"value"` // 0.1 ~ 5
}

// 锐化
type Sharpen struct {
	Value float64 `json:"value"` // 0.1 ~ 10
}

// 高斯模糊
type Gaussblur struct {
	Value float64 `json:"value"`
}

type Image struct {
	origin     string      `json:"-"`          // 缓存原文件名
	Src        string      `json:"src"`        // 图片地址
	Crop       *Crop       `json:"crop"`       // 裁剪,上右下左各裁了多少,最终应该是按百分比计算
	Flip       *Flip       `json:"flip"`       // 翻转,水平和垂直
	Affine     *Affine     `json:"affine"`     // 仿射变换
	Mapim      *Mapim      `json:"mapim"`      // 透视变换
	Resize     *Resize     `json:"resize"`     // 缩放,从中心点按倍数缩放
	Rotation   *Rotation   `json:"rotation"`   // 旋转,按角度
	Linear     *Linear     `json:"linear"`     // 亮度调节
	Saturation *Saturation `json:"saturation"` // 饱和度
	Gamma      *Gamma      `json:"gamma"`      // Gamma校正
	Sharpen    *Sharpen    `json:"sharpen"`    // 锐化
	Gaussblur  *Gaussblur  `json:"gaussblur"`  // 高斯模糊
}

type ImageMeta struct {
	Width  int
	Height int
}

type Processor interface {
	output(input, output string) error
}

func vipsheader(path string) (*ImageMeta, error) {
	out, _, err := funcs.RunCommand(true, libvips.HeaderBin(), path)
	if err != nil {
		return nil, err
	}

	// 解析：1920x1080
	re := regexp.MustCompile(`(\d+)x(\d+)`)
	match := re.FindStringSubmatch(out)
	if len(match) != 3 {
		return nil, fmt.Errorf("failed to parse size: %s", string(out))
	}

	w, _ := strconv.Atoi(match[1])
	h, _ := strconv.Atoi(match[2])

	return &ImageMeta{
		Width:  w,
		Height: h,
	}, nil
}

func runVips(args ...string) (str string, err error) {
	str, _, err = funcs.RunCommand(true, libvips.Bin(), args...)
	return
}
