# Notes on this kata

This is the craziest I've seen so far

What challenged me the most here was working with generics

I will implement two versions of this to compare performance:
- A less performant one making all the mistakes in the instructions 

## cpu2.prof notes

Benchmark result:
- `BenchmarkShardedMapParallel-8`: `53416594` iterations, `112.4 ns/op`, `12 B/op`, `1 allocs/op`

Profile summary:
- `getShard` is a major hotspot because it still uses `fmt.Sprint` for generic hashing.
- `shard.Set` and `shard.Get` show up heavily, with lock contention visible in `sync.(*RWMutex).Lock` and `sync.(*Mutex).Lock`.
- `fmt.Sprint`, `fmt.(*pp).doPrint`, and `fmt.(*fmt).fmtInteger` contribute a lot of overhead and allocations.
- Runtime scheduling and contention are also visible, which is expected under `RunParallel`.

Takeaway:
- The benchmark is working, but the hashing path is too expensive for a high-throughput sharded map.
- Next optimization to try: replace `fmt.Sprint` with type-specific hashing/encoding.
