# When to use percpu

**Do not** use this module if you do not understand and fully accept the
[caveats](README.md#important-caveats) listed in the README.

Percpu allows the programmer to shard a data type so that each CPU has a local
shard. When code uses the data type, it can access the shard that is local to
the CPU on which the goroutine is running.

(This is a simplification. In fact, the underlying entity is not really a CPU,
but a logical processor -- a **P**, in the parlance of the Go runtime.)

This is a useful mechanism for a specific sort of scenario:

* Many goroutines need to access a piece of shared state
* That shared state may be sharded into independent values
* The access entails high-frequency mutation of that shared state -- typically
  many thousands of writes per second or more
* There is no straightforward way to give each goroutine its own shard

A classic example is a counter. Suppose many goroutines need to increment a
counter at a high rate. An increment of an integer is very cheap, so as the
workload scales up to multiple CPUs, the throughput will be completely
bottlenecked on contention of the shared state (whether that is a mutex or
whether the increments use atomics).

You can split up the counter into several different integers which may be
incremented independently and then summed at the end.

**Ideally, you should structure the code such that each goroutine has its own
goroutine-local counter.** Then each goroutine can compute its own sum with zero
contention and no synchronization at all. At the end, you can sum them all up.

However, there are sometimes scenarios that don't lend themselves to this kind
of design. For example, the goroutines may not be under the control of the code
that runs on them and increments the counter (consider a high-performance server
where the goroutines are created by the listening loop).

In this situation, percpu might be useful. It allows the counter to be sharded
by CPU and then for each CPU-local counter to be incremented by each goroutine.
This avoids all the contention without having to associate each counter shard
with the lifetime of a goroutine.

Such a counter is implemented in this package as `Counter`. Unlike a mutex or a
single atomic variable, a `percpu.Counter` scales linearly to any number of CPUs.

## When not to use percpu

In contrast, let us consider scenarios where percpu probably doesn't make sense.

If the data type you're considering doesn't lend itself to being sharded, percpu
is not applicable.

If you have a concurrent operation, but the rate at which you need to run it is
1000 times per second, percpu is not needed. You can simply use a mutex or
atomics; the contention will not be an issue.

(By "contention", I'm referring to CPU cache contention on a shared variable.
If your code takes a lock and then does something slow, such as talking to a
database, then clearly the lock contention could severely bottleneck the program.
But percpu is not the solution to that kind of problem -- instead, the structure
of the code should be changed to avoid holding the lock for a long time.)

If your program creates N worker goroutines to do a task, runs them all, and
then waits at the end, you don't need percpu -- simply give each worker
goroutine a shard of the shared value.

If your workload is mostly reads and only occasional writes, percpu is probably
not applicable (and you probably don't need to shard your data type at all).
There are certainly [pitfalls] with this kind of workload if you rely on
mutexes, but they should be avoidable with careful use of atomics; the problem
is not the cache contention issue that percpu addresses.

[pitfalls]: https://blog.nelhage.com/post/rwlock-contention/
