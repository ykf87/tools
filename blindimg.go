package main

import (
	"fmt"
	"tools/runtimes/blindwm"
)

func Maker() {
	input := "data/keyframe_001.jpg"
	output := "data/001.jpg"
	txt := "sdfsgggg"

	fmt.Println(input, output, txt)

	str, err := blindwm.DecodeImageTxt(output, 64)
	fmt.Println(str, err)

	// length, err := blindwm.AddImageTxt(input, output, txt)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(length)
}
