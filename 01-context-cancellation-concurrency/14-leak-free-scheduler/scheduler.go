package main

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type Cache struct {
	mu    sync.Mutex
	cache map[string]int
}


func NewCache() *Cache {
	return &Cache{
		cache: make(map[string]int),
	}
}

type Scheduler struct {
	cache  *Cache
	logger *slog.Logger
}

func NewScheduler(c *Cache, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		cache:  c,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (s *Scheduler) Run(ctx context.Context, job func(context.Context) error) error {
	jitter := time.Duration(rand.IntN(500)) * time.Millisecond
	duration := (5 * time.Second) + time.Duration(jitter)
	ticker := time.NewTicker(duration)
	s.logger.Info("Performing job with duration", "duration", duration.String())
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				if err := job(ctx); err != nil {
					s.logger.Info("Cache refreshed at", "time", time.Now().String())
					return err
				}
			}
		}
	})

	return g.Wait()
}

func main() {
	ctx := context.Background()
	cache := NewCache()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	refresher := NewScheduler(cache, logger)
	job := func(ctx context.Context) error {
		// Simulate putting something in the cache
		num := rand.IntN(100)
		cache.mu.Lock()
		cache.cache["entry"] = num
		logger.Info("Refreshed with", "value", num)
		cache.mu.Unlock()
		return nil
	}

	err := refresher.Run(ctx, job)
	if err != nil {
		logger.Error("something went wrong while refreshing the cache", "error", err.Error())
	}
}
