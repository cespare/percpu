package percpu

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

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
