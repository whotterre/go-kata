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

## Implementation note — type conflict and fix

- Problem: I hit a build error caused by duplicate `shard` type declarations (one in the `slow/` variant and one in the main package). Go compiles all files in the same package directory, so identical type names caused conflicting declarations.
- Fix applied: introduced a local `fastShard` type and `FastShardedMap` in `fastMap.go` (instead of reusing `shard`/`ShardedMap`), and updated tests/benchmarks to exercise the fast implementation.

## Benchmark comparison (before → after)

- Slow (original): `BenchmarkShardedMapParallel-8` — 112.4 ns/op, 12 B/op, 1 allocs/op
- Fast (type-specific hashing): `BenchmarkShardedMapParallel-8` — 88.15 ns/op, 6 B/op, 0 allocs/op

Notes:
- The improvement comes mainly from removing `fmt.Sprint` allocations on the hot path and using a small stack buffer + `binary.LittleEndian` for integer keys.
- Next actions: re-run CPU profile on the fast map, consider alternative lightweight integer mixers to remove `fnv` overhead.
