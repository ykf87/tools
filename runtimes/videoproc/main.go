package videoproc

import "tools/runtimes/mainsignal"

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
