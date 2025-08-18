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
		thr.Push(strconv.Itoa(i))
	}
	time.Sleep(10 * time.Second)
}


func TestThrottle(t *testing.T) {
	start := time.Now()
	thr := NewThrottler(func(key string, payload []any) {
		fmt.Println("KEY", time.Since(start), payload)
	}, 3000, false)

	for i := 0; i < 1000; i++ {
		time.Sleep(10 * time.Millisecond)
		thr.Push("a", strconv.Itoa(i))
	}
	time.Sleep(10 * time.Second)
}
