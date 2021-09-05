package clrand

import (
	stdrand "math/rand"
	"testing"

	"golang.org/x/exp/rand"
)

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
