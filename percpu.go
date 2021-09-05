// Package percpu provides best-effort CPU-local sharded values.
package percpu

import (
	"runtime"
	_ "unsafe"
)

// Values is a sharded set of values which have an affinity for a particular
// processor. This can be used to avoid cache contention when updating a shared
// value simultaneously from many goroutines.
type Values struct {
	shards []interface{}
}

// NewValues constructs a Values using the provided constructor function to
// create each shard value.
func NewValues(newVal func() interface{}) *Values {
	shards := make([]interface{}, runtime.GOMAXPROCS(0))
	for i := range shards {
		shards[i] = newVal()
	}
	return &Values{shards: shards}
}

// Get returns one of the values in vs.
//
// The value tends to be the one associated with the current processor.
// However, goroutines can migrate at any time, and it may be the case
// that a different goroutine is accessing the same value concurrently.
// All access of the returned value must use further synchronization
// mechanisms.
//
// BUG(cespare): If GOMAXPROCS has changed since a Values was created with
// NewValues, Get may panic.
func (vs *Values) Get() interface{} {
	return vs.shards[getProcID()]
}

// Do runs fn on all of the values in vs.
func (vs *Values) Do(fn func(interface{})) {
	for _, pv := range vs.shards {
		fn(pv)
	}
}

//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin() int

func getProcID() int {
	pid := runtime_procPin()
	runtime_procUnpin()
	return pid
}
