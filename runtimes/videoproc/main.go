package videoproc

import (
	"tools/runtimes/mainsignal"
)

var limit = make(chan byte, 1)

func init() {
	go func() {
		for {
			select {
			case <-limit:
				continue
			case <-mainsignal.MainCtx.Done():
				return
			}
		}
	}()
}

func SecMaker(videos, audios []string) (*Maker, error) {
	mk := new(Maker)
	mk.Srcs = videos
	mk.Audios = audios
	mk.Factory = new(Factory)
	return mk, nil
}
