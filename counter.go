package percpu

import (
	"sync/atomic"
)

// A Counter is an int64 counter which may be efficiently incremented
// by many goroutines concurrently.
type Counter struct {
	vs *Values
}

type cval struct {
	n   int64
	pad [120]byte // prevent false sharing
}

// NewCounter returns a fresh Counter initialized to zero.
func NewCounter() *Counter {
	vs := NewValues(func() interface{} { return new(cval) })
	return &Counter{vs: vs}
}

// Add adds n to the total count.
func (c *Counter) Add(n int64) {
	atomic.AddInt64(&c.vs.Get().(*cval).n, n)
}

// Load returns the current total counter value.
func (c *Counter) Load() int64 {
	var sum int64
	c.vs.Do(func(v interface{}) {
		sum += atomic.LoadInt64(&v.(*cval).n)
	})
	return sum
}
