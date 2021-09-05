// Package clrand provides CPU-local random number generators.
package clrand

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/cespare/percpu"
	"golang.org/x/exp/rand"
)

// A Source is a source of uniformly-distributed pseudo-random uint64 values
// in the range [0, 1<<64). It implements golang.org/x/exp/rand.Source.
type Source struct {
	vs *percpu.Values // of *sval
}

type lockedPCGSource struct {
	mu  sync.Mutex
	pcg rand.PCGSource
}

type sval struct {
	lockedPCGSource
	pad [128 - unsafe.Sizeof(lockedPCGSource{})%128]byte
}

func NewSource(seed uint64) *Source {
	vs := percpu.NewValues(func() interface{} {
		var sv sval
		sv.pcg.Seed(atomic.AddUint64(&seed, 1) - 1)
		return &sv
	})
	return &Source{vs: vs}
}

func (s *Source) Seed(seed uint64) {
	s.vs.Do(func(v interface{}) {
		sv := v.(*sval)
		sv.mu.Lock()
		defer sv.mu.Unlock()
		sv.pcg.Seed(atomic.AddUint64(&seed, 1) - 1)
	})
}

func (s *Source) Uint64() uint64 {
	sv := s.vs.Get().(*sval)
	sv.mu.Lock()
	defer sv.mu.Unlock()
	return sv.pcg.Uint64()
}

var globalRand *rand.Rand

func init() {
	var b [8]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		panic(err)
	}
	seed := binary.BigEndian.Uint64(b[:])
	globalRand = rand.New(NewSource(seed))
}

// ExpFloat64 returns an exponentially distributed float64 in the range
// (0, +math.MaxFloat64] with an exponential distribution whose rate parameter
// (lambda) is 1 and whose mean is 1/lambda (1).
// To produce a distribution with a different rate parameter,
// callers can adjust the output using:
//
//  sample = ExpFloat64() / desiredRateParameter
//
func ExpFloat64() float64 { return globalRand.ExpFloat64() }

// Float32 returns, as a float32, a pseudo-random number in [0.0,1.0).
func Float32() float32 { return globalRand.Float32() }

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
func Float64() float64 { return globalRand.Float64() }

// Int returns a non-negative pseudo-random int.
func Int() int { return globalRand.Int() }

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.
func Int31() int32 { return globalRand.Int31() }

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Int31n(n int32) int32 { return globalRand.Int31n(n) }

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int63() int64 { return globalRand.Int63() }

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Int63n(n int64) int64 { return globalRand.Int63n(n) }

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Intn(n int) int { return globalRand.Intn(n) }

// NormFloat64 returns a normally distributed float64 in the range
// [-math.MaxFloat64, +math.MaxFloat64] with
// standard normal distribution (mean = 0, stddev = 1).
// To produce a different normal distribution, callers can
// adjust the output using:
//
//  sample = NormFloat64() * desiredStdDev + desiredMean
//
func NormFloat64() float64 { return globalRand.NormFloat64() }

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers [0,n).
func Perm(n int) []int { return globalRand.Perm(n) }

// Read generates len(p) random bytes and writes them into p. It
// always returns len(p) and a nil error.
func Read(p []byte) (n int, err error) { return globalRand.Read(p) }

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) { globalRand.Shuffle(n, swap) }

// Uint32 returns a pseudo-random 32-bit value as a uint32.
func Uint32() uint32 { return globalRand.Uint32() }

// Uint64 returns a pseudo-random 64-bit integer as a uint64.
func Uint64() uint64 { return globalRand.Uint64() }
