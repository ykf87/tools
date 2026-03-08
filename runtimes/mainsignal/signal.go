package mainsignal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// var Quit chan os.Signal
var MainCtx context.Context
var MainStop context.CancelFunc
var MainWait sync.WaitGroup

func init() {
	MainCtx, MainStop = signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
}
