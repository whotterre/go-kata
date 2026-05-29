package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

// Less performant / initial effort nooby way of doing this

type shard[K comparable, V any] struct {
	mu sync.RWMutex
	m map[K]V
	hasher func(k string) uint64
}

type ShardedMap[K comparable, V any] struct {
	shards []shard[K, V]
}

func NewShardedMap[K comparable, V any](numShards int) *ShardedMap[K, V] {
	if numShards <= 0 {
		numShards = 16
	}
	s := &ShardedMap[K, V]{
		shards: make([]shard[K, V], numShards),
	}
	for i := range s.shards {
		s.shards[i].m = make(map[K]V)
	}
	return s
}


func (s *ShardedMap[K, V]) getShard(key K) *shard[K, V] {
	if len(s.shards) == 0 {
		return nil
	}
	var data []byte
	switch v := any(key).(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data = []byte(fmt.Sprint(v))
	}

	h := fnv.New64a()
	_, _ = h.Write(data)
	hash := h.Sum64()
	idx := int(hash % uint64(len(s.shards)))
	return &s.shards[idx]
}

func(s *shard[K, V]) Get(key K) (V, bool){
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.m[key]
	if !ok {
		var zero V
		return zero, false
	}
	return value, true
}
func(s *shard[K, V]) Set(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}
func(s *shard[K, V]) Delete(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}
func(s *shard[K, V]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}
