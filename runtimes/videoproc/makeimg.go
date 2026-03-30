// 创建图片,用于补帧
package videoproc

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
)

func CreateSolidImage(path string, width, height int, c color.RGBA) error {
	if width <= 0 || height <= 0 {
		return errors.New("invalid size")
	}

	// 1. 创建图像
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 2. 构造一行像素（关键优化点）
	row := make([]byte, width*4)
	for i := 0; i < len(row); i += 4 {
		row[i] = c.R
		row[i+1] = c.G
		row[i+2] = c.B
		row[i+3] = c.A
	}

	// 3. 批量填充（最快方式）
	for y := range height {
		start := y * img.Stride
		copy(img.Pix[start:start+width*4], row)
	}

	// 4. 创建文件
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// 5. 根据后缀编码
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".png":
		return png.Encode(f, img)

	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{
			Quality: 90,
		})

	case ".bmp":
		return bmp.Encode(f, img)

	default:
		return errors.New("unsupported format: " + ext)
	}
}
