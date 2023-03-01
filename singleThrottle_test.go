package throttle

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSingleThrottle(t *testing.T) {
	start := time.Now()
	thr := NewSingleThrottler(func(key []string) {
		fmt.Println("KEY", time.Since(start), strings.Join(key, ";"))
	}, 2000)

	for i := 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		thr.Push(strconv.Itoa(i), nil)
	}
	time.Sleep(10 * time.Second)
}
