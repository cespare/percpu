package percpu

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCounter(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup
	const n = 100
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				c.Add(1)
			}
		}()
	}
	wg.Wait()
	got := c.Load()
	if want := int64(n * n); got != want {
		t.Fatalf("got total %d; want %d", got, want)
	}
}

func TestCounterReset(t *testing.T) {
	c := NewCounter()
	var wg sync.WaitGroup
	const n = 100
	var resetSum int64
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < n; i++ {
				c.Add(1)
				if i%20 == 0 {
					atomic.AddInt64(&resetSum, c.Reset())
				}
			}
		}()
	}
	wg.Wait()
	resetSum += c.Reset()
	if n := c.Load(); n != 0 {
		t.Fatalf("after Reset, Load was %d", n)
	}
	if want := int64(n * n); resetSum != want {
		t.Fatalf("got total Resets=%d; want %d", resetSum, want)
	}
}

// Measure overhead without contention.
func BenchmarkCounterNonParallel(b *testing.B) {
	c := NewCounter()
	for i := 0; i < b.N; i++ {
		c.Add(1)
	}
}

type mutexCounter struct {
	mu sync.Mutex
	n  int64
}

func (c *mutexCounter) Add(n int64) {
	c.mu.Lock()
	c.n += n
	c.mu.Unlock()
}

func (c *mutexCounter) Load() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

type atomicCounter struct {
	n int64
}

func (c *atomicCounter) Add(n int64) {
	atomic.AddInt64(&c.n, n)
}

func (c *atomicCounter) Load() int64 {
	return atomic.LoadInt64(&c.n)
}

func BenchmarkCounterParallel(b *testing.B) {
	type counter interface {
		Add(int64)
		Load() int64
	}
	for _, newFunc := range []func() counter{
		func() counter { return new(mutexCounter) },
		func() counter { return new(atomicCounter) },
		func() counter { return NewCounter() },
	} {
		name := fmt.Sprintf("%T", newFunc())
		if i := strings.LastIndex(name, "."); i >= 0 {
			name = name[i+1:]
		}
		b.Run(name, func(b *testing.B) {
			c := newFunc()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					c.Add(1)
				}
			})
		})
	}
}
