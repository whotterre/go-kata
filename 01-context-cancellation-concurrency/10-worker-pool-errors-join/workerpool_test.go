package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_Run(t *testing.T) {
	t.Run("Processes all jobs successfully", func(t *testing.T) {
		pool := NewPool(3)
		jobCount := 10
		jobs := make(chan Job, jobCount)
		var processed int32

		for range jobCount {
			jobs <- func(ctx context.Context) error {
				atomic.AddInt32(&processed, 1)
				return nil
			}
		}
		close(jobs)

		err := pool.Run(context.Background(), jobs)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if atomic.LoadInt32(&processed) != int32(jobCount) {
			t.Errorf("expected %d jobs processed, got %d", jobCount, processed)
		}
	})

	t.Run("Collects multiple errors when StopOnFirstError is false", func(t *testing.T) {
		pool := NewPool(3)
		pool.StopOnFirstError = false
		jobs := make(chan Job, 3)

		err1 := errors.New("fail 1")
		err2 := errors.New("fail 2")

		jobs <- func(ctx context.Context) error { return err1 }
		jobs <- func(ctx context.Context) error { return err2 }
		jobs <- func(ctx context.Context) error { return nil }
		close(jobs)

		err := pool.Run(context.Background(), jobs)
		if err == nil {
			t.Fatal("expected joined errors, got nil")
		}

		// Verify both errors exist in the joined error
		if !errors.Is(err, err1) || !errors.Is(err, err2) {
			t.Errorf("missing expected errors in aggregate: %v", err)
		}
	})

	t.Run("Fails fast when StopOnFirstError is true", func(t *testing.T) {
		pool := NewPool(3)
		pool.StopOnFirstError = true
		jobs := make(chan Job, 100) // Large buffer to ensure jobs remain

		jobs <- func(ctx context.Context) error {
			return errors.New("boom")
		}

		// Fill remaining with jobs that should be cancelled
		for i := 0; i < 99; i++ {
			jobs <- func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			}
		}
		close(jobs)

		start := time.Now()
		err := pool.Run(context.Background(), jobs)
		duration := time.Since(start)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// If it failed fast, it shouldn't have waited for 99 * 10ms
		if duration > 500*time.Millisecond {
			t.Errorf("took too long to fail: %v", duration)
		}
	})

	t.Run("Respects external context cancellation", func(t *testing.T) {
		pool := NewPool(3)
		jobs := make(chan Job, 10)
		ctx, cancel := context.WithCancel(context.Background())

		// Start a long job
		jobs <- func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}
		
		// Cancel externally
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
			close(jobs)
		}()

		err := pool.Run(ctx, jobs)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}