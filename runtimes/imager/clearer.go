package imager

// import (
// 	"fmt"
// 	"os"
// 	"strconv"
// 	"tools/runtimes/clearer"
// )

// func (c *Clearer) output(img *Image) error {
// 	outname := fmt.Sprintf("%s.png", img.OutFile)
// 	_, err := clearer.Clearers(img.OutFile, outname, "")
// 	if err != nil {
// 		return err
// 	}

// 	_, err = runVips("resize", outname, img.OutFile, strconv.FormatFloat(0.25, 'f', -1, 64))
// 	if err != nil {
// 		return err
// 	}

// 	os.Remove(outname)

// 	return err
// }
