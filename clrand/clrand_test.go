package clrand

import (
	stdrand "math/rand"
	"runtime"
	"sync"
	"testing"

	"golang.org/x/exp/rand"
)

// Confirm that we don't see any duplicates in a few thousand generated values.
func TestSource(t *testing.T) {
	n := runtime.GOMAXPROCS(0)
	allSeen := make([]map[uint64]struct{}, n)
	var wg sync.WaitGroup
	source := NewSource()
	for i := range allSeen {
		seen := make(map[uint64]struct{})
		allSeen[i] = seen
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				seen[source.Uint64()] = struct{}{}
			}
		}()
	}

	wg.Wait()
	seen := make(map[uint64]struct{})
	for _, seen1 := range allSeen {
		for x := range seen1 {
			if _, ok := seen[x]; ok {
				t.Fatalf("saw value %d twice", x)
			}
			seen[x] = struct{}{}
		}
	}
}

func BenchmarkParallel(b *testing.B) {
	for _, tt := range []struct {
		name string
		fn   func() uint64
	}{
		{"math-rand.Uint64", stdrand.Uint64},
		{"exp-rand.Uint64", rand.Uint64},
		{"Uint64", Uint64},
	} {
		b.Run(tt.name, func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					tt.fn()
				}
			})
		})
	}
}
