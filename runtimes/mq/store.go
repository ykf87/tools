package mq

type Store interface {
	Save(topic, payload string) (int64, error)
	LoadPending() ([]*Message, error)
	MarkDone(id int64) error
	MarkFailed(id int64) error
}
