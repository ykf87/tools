package imager

import (
	"errors"
	"os"
	"tools/runtimes/libvips"
	// "github.com/davidbyttow/govips/v2/vips"
	// "github.com/h2non/bimg"
)

func init() {
	libvips.Init()
}

// 如果输出是空的，则覆盖原图
func NewImager(src string) (*Image, error) {
	if _, err := os.Stat(src); err != nil {
		return nil, err
	}
	return &Image{
		Src: src,
	}, nil
}

func (img *Image) copySrc() error {
	newFileName := img.Src + "-inputer.v" //filepath.Join(filepath.Dir(img.Src), strings.ReplaceAll(filepath.Base(img.Src), ".", "_maker."))

	_, err := runVips("copy", img.Src, newFileName)
	if err != nil {
		return err
	}
	// if err := funcs.CopyFile(img.Src, newFileName); err != nil {
	// 	return err
	// }
	img.origin = img.Src
	img.Src = newFileName
	return nil
}

func (img *Image) Output(output string, maxtry int) (err error) {
	if maxtry < 1 {
		maxtry = 1
	}
	for {
		if err := img.run(output); err != nil {
			maxtry--
			if maxtry < 1 {
				return err
			}
			continue
		}
		return nil
	}
}

func (img *Image) run(output string) (err error) {
	if output != "" {
		img.OutFile = output
	}

	if img.OutFile == "" {
		err = errors.New("输出文件夹是空的")
		return
	}

	if err = img.copySrc(); err != nil {
		return
	}
	meta, err := vipsheader(img.Src) // 你自己实现获取宽高
	if err != nil {
		return err
	}
	img.w, img.h = meta.Width, meta.Height
	if img.Width <= 0 {
		img.Width = img.w
	}
	if img.Height <= 0 {
		img.Height = img.h
	}

	// 按顺序执行（顺序很重要）
	steps := img.buildpip()

	img.outtemp = img.OutFile + "-output.v" //filepath.Join(filepath.Dir(output), strings.ReplaceAll(filepath.Base(output), ".", "__outer."))

	for _, step := range steps {
		if err = step.output(img); err != nil {
			os.Remove(img.Src)
			return
		}
		os.Rename(img.outtemp, img.Src)
	}

	if _, err := runVips("copy", img.Src, img.OutFile); err != nil {
		return err
	}
	os.Remove(img.Src)
	img.Src = img.origin
	return
}

func (img *Image) buildpip() []Processor {
	var steps []Processor
	if img.Rotation != nil {
		steps = append(steps, img.Rotation)
	}
	if img.Crop != nil {
		steps = append(steps, img.Crop)
	}
	if img.Flip != nil {
		steps = append(steps, img.Flip)
	}
	if img.Affine != nil {
		steps = append(steps, img.Affine)
	}
	if img.Mapim != nil {
		steps = append(steps, img.Mapim)
	}
	if img.Resize != nil {
		steps = append(steps, img.Resize)
	}
	if img.Linear != nil {
		steps = append(steps, img.Linear)
	}
	if img.Saturation != nil {
		steps = append(steps, img.Saturation)
	}
	if img.Gamma != nil {
		steps = append(steps, img.Gamma)
	}
	if img.Sharpen != nil {
		steps = append(steps, img.Sharpen)
	}
	if img.Gaussblur != nil {
		steps = append(steps, img.Gaussblur)
	}
	if img.KeepWH != nil {
		steps = append(steps, img.KeepWH)
	}
	// if img.Clearer != nil {
	// 	steps = append(steps, img.Clearer)
	// }
	return steps
}
