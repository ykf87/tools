package mainsignal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// var Quit chan os.Signal
var MainCtx context.Context
var MainStop context.CancelFunc

func init() {
	MainCtx, MainStop = signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	// Quit = make(chan os.Signal, 1)
	// signal.Notify(Quit, syscall.SIGINT, syscall.SIGTERM)
}
