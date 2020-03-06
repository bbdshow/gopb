package timing

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func randSleep() {
	n := rand.Int63n(1000)
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func TestTiming(t *testing.T) {
	timing := NewTiming(true)
	wg := sync.WaitGroup{}
	count := 100
	concurrency := 3
	for concurrency > 0 {
		concurrency--

		wg.Add(1)
		go func(con, c int) {
			for c > 0 {
				c--
				timing.Do(strconv.Itoa(con), func() {
					randSleep()
				})
			}
			wg.Done()
		}(concurrency, count)
	}

	wg.Wait()

	data := timing.GetMethodData()
	for _, d := range data {
		t.Log(d)
	}
}
