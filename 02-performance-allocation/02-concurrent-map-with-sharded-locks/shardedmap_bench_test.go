package main

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkShardedMapParallel(b *testing.B) {
	m := NewShardedMap[int, string](64)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			k := r.Int()
			sh := m.getShard(k % 1024)
			sh.Set(k%1024, "v")
			sh.Get(k % 1024)
		}
	})
}
