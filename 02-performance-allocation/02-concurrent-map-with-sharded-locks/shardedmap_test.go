package main

import (
	"math/rand"
	"sync"
	"testing"
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
	shardedMap := NewShardedMap[int, string](1)

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
	shardedMap := NewShardedMap[int, string](64)

	value := genRandomString(8)
	wg.Add(n)
	for i := range n {
		i := i
		go func(i int) {
			defer wg.Done()
			shard := shardedMap.getShard(i)
			shard.Set(i, value)
		}(i)
	}
	wg.Wait()

}
