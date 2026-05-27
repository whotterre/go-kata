package main

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"os"
	"runtime"
	"testing"
	"time"
)

func Test_NoGoroutinesAfterExit(t *testing.T) {
	baseline := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	testCache := NewCache()
	testLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	testScheduler := NewScheduler(testCache, testLogger)
	job := func(ctx context.Context) error {
		num := rand.IntN(100)
		testLogger.Info("Refreshed with", "value", num)
		testCache.Put(num)
		return nil
	}

	done := make(chan error, 10)
	for range 10 {
		go func() {
			done <- testScheduler.Run(ctx, job)
		}()
	}

	time.Sleep(25 * time.Millisecond)
	cancel()

	for range 10 {
		select {
		case err := <-done:
			if err != context.Canceled {
				t.Fatalf("expected context.Canceled, got %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for scheduler to exit")
		}
	}

	after := runtime.NumGoroutine()
	if after > baseline + 5 {
		t.Fatalf("possible goroutine leak: baseline=%d after=%d", baseline, after)
	}
}
