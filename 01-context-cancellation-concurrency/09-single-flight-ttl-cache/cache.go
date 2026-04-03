package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type entry[V any] struct {
	value     V
	createdAt time.Time
}

type Cache[K comparable, V any] struct {
	contents map[K]entry[V]
	mu       sync.RWMutex
	group    singleflight.Group
	ttl      time.Duration
}

func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V]{
	return &Cache[K, V]{
		contents: make(map[K]entry[V]),
		ttl: ttl,
	}
}

func (c *Cache[K, V]) Get(ctx context.Context, key K, loader func(context.Context) (V, error)) (V, error) {
	// Read and check whether key exists in cache
	c.mu.RLock()
	e, isInCache := c.contents[key]
	c.mu.RUnlock()

	if isInCache {
		expiryTime := e.createdAt.Add(c.ttl)
		if time.Now().Before(expiryTime) {
			return e.value, nil
		}
	}

	// refresh the cache
	resultChan := c.group.DoChan(fmt.Sprintf("%v", key), func() (any, error) {
		val, err := loader(ctx)
		if err != nil {
			return nil, fmt.Errorf("loader error for key %v: %w", key, err)
		}

		c.mu.Lock()
		c.contents[key] = entry[V]{
			value:     val,
			createdAt: time.Now(),
		}
		c.mu.Unlock()

		return val, nil
	})

	select {
	// if something goes wrong, and ctx.Done() - return
	case <-ctx.Done():
		var zero V
		return zero, ctx.Err()
	case res := <-resultChan:
		if res.Err != nil {
			var zero V
			return zero, res.Err
		}
		return res.Val.(V), nil
	}

}
