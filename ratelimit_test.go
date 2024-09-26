package throttle

import (
	"fmt"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	for i := 0; i < 100; i++ {
		time.Sleep(1 * time.Second)
		failed := RateLimit("thanh", 10, 100)
		fmt.Println("FF", failed)
	}
}
