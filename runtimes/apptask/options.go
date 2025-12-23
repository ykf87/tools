package apptask

import "time"

type Options struct {
	TickInterval time.Duration
	Store        Store
}
