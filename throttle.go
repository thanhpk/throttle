package throttle

import (
	"sync"
	"time"
)

type Throttler struct {
	*sync.Mutex
	wait     int64 // ms
	cache    map[string][]any
	runningM map[string]bool
	handler  func(string, []any)
	leading  bool
}

func NewThrottler(handler func(string, []any), wait int64, leading bool) *Throttler {
	me := &Throttler{
		Mutex:    &sync.Mutex{},
		wait:     wait / 2,
		runningM: make(map[string]bool),
		cache:    make(map[string][]any),
		handler:  handler,
		leading:  leading,
	}
	return me
}

func (me *Throttler) Push(key string, payload any) {
	me.Lock()
	defer me.Unlock()

	me.cache[key] = append(me.cache[key], payload)
	if me.runningM[key] {
		return
	}
	go me.run(key)
}

func (me *Throttler) run(key string) {
	me.Lock()
	if me.runningM[key] {
		me.Unlock()
		return
	}
	me.runningM[key] = true

	me.Unlock()

	if !me.leading {
		// sleep before, we dont want to call handle immediatly
		time.Sleep(time.Duration(me.wait) * time.Millisecond)
	}

	me.Lock()
	payloads := me.cache[key]
	me.cache[key] = make([]any, 0)
	me.Unlock()

	if len(payloads) > 0 {
		me.handler(key, payloads)
	}

	// sleep after
	time.Sleep(time.Duration(me.wait) * time.Millisecond)

	me.Lock()
	delete(me.runningM, key)
	if len(me.cache[key]) > 0 {
		// there is unfinished job
		go me.run(key)
	}
	me.Unlock()
}

/*

	throttler := NewThrottler(100, true)

	go func() {
		for msgs := range throttler.Message {
			s := ""
			for _, msg := range msgs {
				s += msg.(string) + ","
			}
			fmt.Println("GOT", s)
		}
	}()

	for i := 0; i < 1000; i++ {
		throttler.Push("a", fmt.Sprintf("%d", i))
		if i%2 == 0 {
			time.Sleep(5 * time.Millisecond)
		}
		time.Sleep(5 * time.Millisecond)
	}

throttler.Push("a", "1")
throttler.Push("a", "a")
throttler.Push("a", "b")
throttler.Push("a", "c")
time.Sleep(200 * time.Millisecond)

throttler.Push("a", "d")
throttler.Push("a", "e")
throttler.Push("a", "f")

time.Sleep(20 * time.Second)
*/
