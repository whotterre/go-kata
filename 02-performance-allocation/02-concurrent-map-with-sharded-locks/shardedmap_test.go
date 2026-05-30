package main

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func genRandomString(length int) string {
	charSet := "abcdefghijklmnoqrstuvwxyz12"
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		idx := rand.Intn(len(charSet))
		b[i] = charSet[idx]
	}
	return string(b)
}

func Test_NoLockContentionOneShard(t *testing.T) {
	var wg sync.WaitGroup
	n := 8
	shardedMap := NewFastShardedMap[int, string](1)

	value := genRandomString(8)
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func(i int) {
			defer wg.Done()
			shard := shardedMap.getShard(i)
			shard.Set(i, value)
		}(i)
	}
	wg.Wait()

}

func Test_NoLockContention64Shards(t *testing.T) {
	var wg sync.WaitGroup
	n := 8
	shardedMap := NewFastShardedMap[int, string](64)

	value := genRandomString(8)
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func(i int) {
			defer wg.Done()
			shard := shardedMap.getShard(i)
			shard.Set(i, value)
		}(i)
	}
	wg.Wait()

}

func TestMemoryUsed(t *testing.T) {
	// Memory test: store 1 million int keys with interface{} values.
	// Fail if ShardedMap uses >50MB extra vs baseline map.
	const n = 1_000_000
	const allowedExtra = 50 * (1 << 20) // 50MB

	// helper to measure heap alloc after running f
	memAfter := func(f func()) uint64 {
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		before := m.HeapAlloc
		f()
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		runtime.ReadMemStats(&m)
		return m.HeapAlloc - before
	}

	t.Logf("allocating baseline map with %d entries", n)
	baselineAlloc := memAfter(func() {
		m := make(map[int]interface{}, n)
		for i := 0; i < n; i++ {
			m[i] = struct{}{}
		}
		// keep m alive until after measurement
		_ = m
	})

	t.Logf("baseline HeapAlloc delta: %d bytes", baselineAlloc)

	t.Logf("allocating ShardedMap with %d entries", n)
	shardedAlloc := memAfter(func() {
		sm := NewFastShardedMap[int, interface{}](256)
		for i := 0; i < n; i++ {
			sm.Set(i, struct{}{})
		}
		_ = sm
	})

	t.Logf("sharded HeapAlloc delta: %d bytes", shardedAlloc)

	var extra int64 = int64(shardedAlloc) - int64(baselineAlloc)
	t.Logf("extra bytes: %d", extra)
	if extra > allowedExtra {
		t.Fatalf("sharded map uses too much extra memory: %d > %d", extra, allowedExtra)
	}
}
