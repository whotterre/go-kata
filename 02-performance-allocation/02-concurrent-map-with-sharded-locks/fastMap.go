package main

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sync"
)

type fastShard[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func (s *fastShard[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	return v, ok
}

func (s *fastShard[K, V]) Set(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *fastShard[K, V]) Delete(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}

func (s *fastShard[K, V]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

// FastShardedMap is an idiomatic sharded map optimized for common key types.
// It avoids fmt.Sprint on the hot path and uses type-specific hashing for ints.
type FastShardedMap[K comparable, V any] struct {
	shards []fastShard[K, V]
}

// NewFastShardedMap creates and initializes a FastShardedMap.
func NewFastShardedMap[K comparable, V any](numShards int) *FastShardedMap[K, V] {
	if numShards <= 0 {
		numShards = 16
	}
	s := make([]fastShard[K, V], numShards)
	for i := range s {
		s[i].m = make(map[K]V)
	}
	return &FastShardedMap[K, V]{shards: s}
}

// hashKey hashes common key types without unnecessary allocations.
func hashKey[K comparable](key K) uint64 {
	switch v := any(key).(type) {
	case string:
		h := fnv.New64a()
		_, _ = h.Write([]byte(v))
		return h.Sum64()
	case []byte:
		h := fnv.New64a()
		_, _ = h.Write(v)
		return h.Sum64()
	case int:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(v))
		h := fnv.New64a()
		_, _ = h.Write(buf[:])
		return h.Sum64()
	case int64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(v))
		h := fnv.New64a()
		_, _ = h.Write(buf[:])
		return h.Sum64()
	case uint64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], v)
		h := fnv.New64a()
		_, _ = h.Write(buf[:])
		return h.Sum64()
	case uint32:
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], v)
		h := fnv.New64a()
		_, _ = h.Write(buf[:])
		return h.Sum64()
	default:
		h := fnv.New64a()
		_, _ = fmt.Fprint(h, v)
		return h.Sum64()
	}
}

func (m *FastShardedMap[K, V]) getShard(key K) *fastShard[K, V] {
	if len(m.shards) == 0 {
		return nil
	}
	idx := int(hashKey(key) % uint64(len(m.shards)))
	return &m.shards[idx]
}

// Convenience methods delegate to the underlying shard with correct locking.
func (m *FastShardedMap[K, V]) Get(key K) (V, bool) { return m.getShard(key).Get(key) }
func (m *FastShardedMap[K, V]) Set(key K, value V)  { m.getShard(key).Set(key, value) }
func (m *FastShardedMap[K, V]) Delete(key K)        { m.getShard(key).Delete(key) }
func (m *FastShardedMap[K, V]) Keys() []K {
	// Collect keys from all shards.
	var all []K
	for i := range m.shards {
		ks := m.shards[i].Keys()
		all = append(all, ks...)
	}
	return all
}
