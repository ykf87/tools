package wmimage

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"strings"

	"github.com/allanpk716/blind-watermark-go/imgutil"
	"github.com/allanpk716/blind-watermark-go/watermark"
	"golang.org/x/image/bmp"
)

func WMImage(src, text string) error {
	if src == "" {
		return errors.New("Empty")
	}
	if text == "" {
		text = "kktoyfun"
	}
	img, err := watermark.LoadImage(src)
	if err != nil {
		return err
	}
	result, err := watermark.EmbedText(img, text, watermark.ModeFast, 0)
	if err != nil {
		return err
	}

	opts := strings.Split(src, ".")
	ext := opts[len(opts)-1]
	ext = strings.ToLower(ext)
	switch ext {
	case "png":
		output := src + ".png"
		err := imgutil.WritePNG(output, result, 0)
		if err != nil {
			return nil
		}
		os.Remove(src)
		os.Rename(output, src)
	case "jpg", "jpeg":
		output := src + ".jpg"
		err := imgutil.WriteJPEG(output, result, 0)
		if err != nil {
			return nil
		}
		os.Remove(src)
		os.Rename(output, src)
	case "bmp":
		output := src + ".png"
		if err := imgutil.WritePNG(output, result, 0); err != nil {
			return err
		}

		tmpn := src + ".tmp"
		if err := os.Rename(src, tmpn); err == nil {
			imgFile, err := os.Open(output)
			if err != nil {
				fmt.Println("打开文件失败:", err, output)
			}
			img, err := png.Decode(imgFile)
			if err != nil {
				fmt.Println("decode文件失败:", err, tmpn)
			}
			imgFile.Close()

			out, err := os.Create(src)
			if err != nil {
				fmt.Println("创建文件失败:", err, src)
			}
			defer out.Close()
			if err := bmp.Encode(out, img); err != nil {
				fmt.Println("生成失败")
				os.Rename(tmpn, src)
			} else {
				os.Remove(tmpn)
				os.Remove(output)
			}
		} else {
			fmt.Println("重命名失败!....")
		}
	default:
		return fmt.Errorf("格式不支持:%s", ext)
	}
	return nil
}
