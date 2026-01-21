package scheduler

type Options struct {
	MaxConcurrency int
	MaxQueueSize   int
	PersistFile    string
}

func DefaultOptions() Options {
	return Options{
		MaxConcurrency: 5,
		MaxQueueSize:   1000,
	}
}
