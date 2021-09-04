# percpu

[![Go Reference](https://pkg.go.dev/badge/github.com/cespare/percpu.svg)](https://pkg.go.dev/github.com/cespare/percpu)

Percpu is a Go package to support best-effort per-CPU sharded values.

See https://github.com/golang/go/issues/18802 for background information.

**IMPORTANT CAVEATS:**

* This package uses `go:linkname` to access unexported functions from inside the
  Go runtime. Those could be changed or removed in a future Go version, breaking
  this package.
* The code in this package assumes that `GOMAXPROCS` does not change. If the
  value of `GOMAXPROCS` changes after creating a `Values` (due to a call to
  `runtime.GOMAXPROCS` with a positive argument), then `Values.Get` may panic.
