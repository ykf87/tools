package imager

import (
	"tools/runtimes/libvips"
	// "github.com/davidbyttow/govips/v2/vips"
	// "github.com/h2non/bimg"
)

func init() {
	libvips.Init()
}

// 如果输出是空的，则覆盖原图
func NewImager(src string) *Image {
	return &Image{
		Src: src,
	}
}

func (img *Image) Output(output string) (err error) {
	// vips.Startup(nil)
	// 按顺序执行（顺序很重要）
	steps := img.buildpip()

	for _, step := range steps {
		if err = step.output(img.Src, output); err != nil {
			return err
		}
	}
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
	return steps
}
