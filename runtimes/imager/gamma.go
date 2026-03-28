package imager

import "path/filepath"

func (g *Gamma) output(input, output string) error {
	tmp1 := filepath.Dir(input) + "/121_linear.jpg"
	runVips("linear", input, tmp1, "1.2", "10")
	// // 1. 转 float
	// if _, err := runVips("cast", input, tmp1, "float"); err != nil {
	// 	return err
	// }
	// // 1. 转 srgb
	// if _, err := runVips("colourspace", tmp1, tmp1, "srgb"); err != nil {
	// 	return err
	// }

	// // 2. 去 alpha
	// if _, err := runVips("extract_band", tmp1, tmp1, "0", "--n", "3"); err != nil {
	// 	return err
	// }

	// if _, err := runVips("colourspace", input, output, "srgb"); err != nil {
	// 	return err
	// }
	// input = output

	// if _, err := runVips("extract_band", input, output, "0", "--n", "3"); err != nil {
	// 	return err
	// }
	// _, err := runVips("gamma", input, output,
	// 	strconv.FormatFloat(g.Value, 'f', -1, 64),
	// )
	// return err
	return nil
}
