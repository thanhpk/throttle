package throttle

import (
	"sync"
	"time"
)

type SingleThrottler struct {
	*sync.Mutex
	wait      int64
	handler   func([]string)
	lastfired int64
	cache     map[string]bool
}

func NewSingleThrottler(handler func([]string), wait int64) *SingleThrottler {
	me := &SingleThrottler{
		Mutex:     &sync.Mutex{},
		wait:      wait, // ms
		cache:     map[string]bool{},
		handler:   handler,
		lastfired: 0,
	}
	go func() {
		for {
			time.Sleep(time.Duration(wait) * time.Millisecond)
			me.Push(SIGNATURE, nil)
		}
	}()
	return me
}

const SIGNATURE = "LKSJDF$&#$%)($##LKJDFLSKJ#)($!)#*@!)#*LJDSF"

func (me *SingleThrottler) Push(key string, i interface{}) {
	me.Lock()
	defer me.Unlock()

	if key != SIGNATURE {
		me.cache[key] = true
	}

	now := time.Now().UnixMilli()
	if now-me.lastfired < me.wait || len(me.cache) == 0 {
		return
	}
	keys := []string{}
	for k := range me.cache {
		keys = append(keys, k)
	}

	me.lastfired = now
	me.cache = map[string]bool{}
	go me.handler(keys)
}
