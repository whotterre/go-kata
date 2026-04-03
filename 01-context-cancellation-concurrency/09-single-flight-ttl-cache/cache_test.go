package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStampede(t *testing.T) {
	c := NewCache[string, int](time.Minute)
	calls := atomic.Int32{}

	loader := func(ctx context.Context) (int, error) {
		calls.Add(1)
		time.Sleep(100 * time.Millisecond) // Simulate slow work
		return 42, nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Get(context.Background(), "hot-key", loader)
		}()
	}
	wg.Wait()

	if calls.Load() != 1 {
		t.Errorf("Expected 1 loader call, got %d", calls.Load())
	}
}