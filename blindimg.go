package main
import(
	"tools/runtimes/blindwm"
	"image"
    "image/png"
    "os"
    "fmt"
)

func Maker(){
	imgsrc := "data/media/6978706990021.png"
	outImg := "data/6978706990021_wm.png"

	// addTxt(imgsrc, outImg, "order=123456;uid=9988;ts=1700000000")
	fmt.Println("水印嵌入完成:", outImg, imgsrc)

	data, err := getImgTxt(outImg)
	if err != nil{
		fmt.Println(err)
		return
	}

	// fmt.Printf("RAW HEX: % x\n", data, imgsrc)
	fmt.Println(string(data),"-----")
}

func addTxt(imgsrc, outImg, txt string) error{
	// 1. 打开图片
	f, err := os.Open(imgsrc)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	// 2. 要嵌入的水印内容
	payload := []byte(txt)

	// 3. 水印参数
	// opt := blindwm.Options{
	// 	Redundancy: 5,     // 冗余倍数
	// 	Strength:   0.02,  // 嵌入强度
	// }

	// 4. 嵌入水印
	wmImg, err := blindwm.EmbedImage(img, payload)
	if err != nil {
		return err
	}

	// 5. 保存新图片
	out, err := os.Create(outImg)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := png.Encode(out, wmImg); err != nil {
		return err
	}
	return nil
}

func getImgTxt(imgsrc string)([]byte, error){
	// ======================
	// 6. 反向提取水印
	// ======================
	// opt := blindwm.Options{
	// 	Redundancy: 5,     // 冗余倍数
	// 	Strength:   0.02,  // 嵌入强度
	// }

	f2, err := os.Open(imgsrc)
	if err != nil {
		return nil, err
	}
	defer f2.Close()

	img2, _, err := image.Decode(f2)
	if err != nil {
		return nil, err
	}

	return blindwm.ExtractImage(img2)
}
