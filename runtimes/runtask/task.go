package runtask

import "sync"

type RunTask struct {
	mu   sync.RWMutex
	wake chan struct{}
}
