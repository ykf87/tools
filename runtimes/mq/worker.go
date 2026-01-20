package mq

import "time"

func (mq *MQ) consumeLoop(ch chan *Message, handler Handler) {
	defer mq.wg.Done()

	for {
		select {
		case <-mq.ctx.Done():
			return
		case msg := <-ch:
			err := handler(mq.ctx, msg)
			if err != nil {
				msg.Retry++
				if msg.Retry > mq.maxRetry {
					_ = mq.store.MarkFailed(msg.ID)
					continue
				}

				go func(m *Message) {
					time.Sleep(time.Second * time.Duration(m.Retry))
					ch <- m
				}(msg)

				continue
			}

			_ = mq.store.MarkDone(msg.ID)
		}
	}
}
