package throttle

import (
	"sync"
	"time"
)

type Message struct {
	Lock  *sync.Mutex
	Value any
}

type SingleSyncThrottler struct {
	*sync.Mutex
	messages []*Message
}

func NewSingleSyncThrottler(handler func([]any), waitMs int64) *SingleSyncThrottler {
	me := &SingleSyncThrottler{Mutex: &sync.Mutex{}}
	go func() {
		for {
			time.Sleep(time.Duration(waitMs) * time.Millisecond)
			me.Lock()
			messages := me.messages
			me.messages = []*Message{}
			me.Unlock()

			values := []any{}
			for _, msg := range messages {
				values = append(values, msg.Value)
			}

			handler(values)
			for _, msg := range messages {
				msg.Lock.Unlock()
			}
		}
	}()
	return me
}

func (me *SingleSyncThrottler) Push(value any) {
	lock := &sync.Mutex{}

	me.Lock()
	me.messages = append(me.messages, &Message{
		Lock:  lock,
		Value: value,
	})

	me.Unlock()
	lock.Lock()
	lock.Lock() // intentionally wait for next batch
	lock.Unlock()
}
