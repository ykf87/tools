package imager

import "tools/runtimes/clearer"

func (c *Clearer) output(img *Image) error {
	_, err := clearer.Clearers(img.Src, img.outtemp, "")
	if err != nil {
		return err
	}
	return nil
}
