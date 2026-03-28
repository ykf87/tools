package imager

import "strconv"

func (g *Gamma) output(input, output string) error {
	tmp1 := input + "_srgb.png"
	tmp2 := input + "_rgb.png"

	// 1. 转 srgb
	if _, err := runVips("colourspace", input, tmp1, "srgb"); err != nil {
		return err
	}

	// 2. 去 alpha
	if _, err := runVips("extract_band", tmp1, tmp2, "0", "--n", "3"); err != nil {
		return err
	}

	// if _, err := runVips("colourspace", input, output, "srgb"); err != nil {
	// 	return err
	// }
	// input = output

	// if _, err := runVips("extract_band", input, output, "0", "--n", "3"); err != nil {
	// 	return err
	// }
	_, err := runVips("gamma", tmp2, output,
		strconv.FormatFloat(g.Value, 'f', -1, 64),
	)
	return err
}
