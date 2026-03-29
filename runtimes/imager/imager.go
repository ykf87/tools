package imager

import (
	"os"
	"path/filepath"
	"strings"
	"tools/runtimes/funcs"
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
	newFileName := filepath.Join(filepath.Dir(img.Src), strings.ReplaceAll(filepath.Base(img.Src), ".", "_maker."))

	if err := funcs.CopyFile(img.Src, newFileName); err != nil {
		return err
	}
	img.origin = img.Src
	img.Src = newFileName
	return nil
}

func (img *Image) Output(output string) (err error) {
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

	// vips.Startup(nil)
	// 按顺序执行（顺序很重要）
	steps := img.buildpip()

	img.outtemp = filepath.Join(filepath.Dir(output), strings.ReplaceAll(filepath.Base(output), ".", "__outer."))

	for _, step := range steps {
		if err = step.output(img); err != nil {
			os.Remove(img.Src)
			return
		}
		os.Rename(img.outtemp, img.Src)
	}

	os.Rename(img.Src, output)
	return

	// if img.Crop != nil {
	// 	if err = img.Crop.apply(img.Src, img.OutputSrc); err != nil {
	// 		return
	// 	}
	// }
	// if img.Gamma != nil {
	// 	img.Gamma.apply(img.Src, img.OutputSrc)
	// }
	// return
	// i := vipscli.NewImage(img.Src)
	// i.SetBinary(libvips.Bin())

	// out, err := i.Process(vipscli.Options{
	// 	Gamma: 1.8,
	// })
	// if err != nil {
	// 	fmt.Println(string(out))
	// 	return err
	// }

	// return os.WriteFile(img.OutputSrc, out, 0644)
	// buffer, err := bimg.Read(img.Src)
	// if err != nil {
	// 	return err
	// }
	// newImage, err := bimg.NewImage(buffer).Process(bimg.Options{
	// 	Gamma: 1.3,
	// })
	// if err != nil {
	// 	return err
	// }

	// return bimg.Write(output, newImage)
	// return nil
}

func (img *Image) buildpip() []Processor {
	var steps []Processor
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
	if img.Rotation != nil {
		steps = append(steps, img.Rotation)
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
	if img.Clearer != nil {
		steps = append(steps, img.Clearer)
	}
	return steps
}
