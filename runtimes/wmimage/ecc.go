package wmimage

import "github.com/klauspost/reedsolomon"

func newRS(data, parity int) reedsolomon.Encoder {
	enc, _ := reedsolomon.New(data, parity)
	return enc
}
