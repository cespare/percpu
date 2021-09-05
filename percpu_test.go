package percpu

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestValues(t *testing.T) {
	numProcs := runtime.GOMAXPROCS(0)
	if numProcs > runtime.NumCPU() {
		t.Skip("unreliable with high GOMAXPROCS")
	}
	// shard -> #goroutines for which the shard is the most frequently seen
	freqCounts := make(map[*int]int)
	// The short timing makes this a little flaky (a goroutine can just
	// finish all its work before another one has started, and reuse the
	// same shard). But we should get a successful attempt pretty quickly.
attemptLoop:
	for attempt := 0; attempt < 10; attempt++ {
		for k := range freqCounts {
			delete(freqCounts, k)
		}
		start := make(chan struct{})
		var wg sync.WaitGroup
		var mu sync.Mutex
		vs := NewValues(func() interface{} { return new(int) })
		for i := 0; i < numProcs; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				seen := make(map[*int]int)
				<-start
				for i := 0; i < 1e6*(attempt+1); i++ {
					p := vs.Get().(*int)
					seen[p]++
				}
				var pmax *int
				max := 0
				for p, n := range seen {
					if n > max {
						pmax = p
						max = n
					}
				}
				mu.Lock()
				freqCounts[pmax]++
				mu.Unlock()
			}()
		}
		close(start)
		wg.Wait()
		for _, n := range freqCounts {
			if n != 1 {
				continue attemptLoop
			}
		}
		return
	}

	t.Fatalf("shards were not handed out evenly to goroutines: %v", freqCounts)
}

// Confirm that getProcID runs and returns the full set of values
// in [0, GOMAXPROCS) as a quick sanity check.
func TestGetProcID(t *testing.T) {
	numProcs := runtime.GOMAXPROCS(0)
	if numProcs > runtime.NumCPU() {
		t.Skip("unreliable with high GOMAXPROCS")
	}
	seen := make([]int64, numProcs)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for range seen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < 1e6; i++ {
				atomic.AddInt64(&seen[getProcID()], 1)
			}
		}()
	}
	close(start)
	wg.Wait()
	for i, n := range seen {
		if n == 0 {
			t.Fatalf("did not see proc id %d (got: %v)", i, seen)
		}
	}
}
