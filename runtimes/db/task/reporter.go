package task

// Reporter 对外通信抽象（日志 / 进度 / WS）
type Reporter interface {
	Log(id int64, msg string)
	Progress(id int64, percent float64)
	Message(id int64, msg string)
}

/************** Multi **************/

type MultiReporter struct {
	list []Reporter
}

func NewMultiReporter(rs ...Reporter) Reporter {
	return &MultiReporter{list: rs}
}

func (m *MultiReporter) Log(id int64, msg string) {
	for _, r := range m.list {
		r.Log(id, msg)
	}
}

func (m *MultiReporter) Progress(id int64, p float64) {
	for _, r := range m.list {
		r.Progress(id, p)
	}
}

func (m *MultiReporter) Message(id int64, msg string) {
	for _, r := range m.list {
		r.Message(id, msg)
	}
}

/************** noop（防 panic） **************/

type noopReporter struct{}

func (n noopReporter) Log(int64, string)       {}
func (n noopReporter) Progress(int64, float64) {}
func (n noopReporter) Message(int64, string)   {}

func safeReporter(r Reporter) Reporter {
	if r == nil {
		return noopReporter{}
	}
	return r
}
