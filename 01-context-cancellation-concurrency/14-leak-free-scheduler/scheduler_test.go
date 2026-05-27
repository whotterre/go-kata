package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"runtime"
	"testing"
	"time"
)

func newTestScheduler() *Scheduler {
	return NewScheduler(NewCache(), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestRunStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := newTestScheduler()
	done := make(chan error, 1)

	go func() {
		done <- scheduler.Run(ctx, func(context.Context) error { return nil })
	}()

	time.Sleep(25 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for scheduler to stop")
	}
}

func TestRunPropagatesContextToJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := newTestScheduler()
	started := make(chan struct{})
	sawCancel := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		done <- scheduler.Run(ctx, func(jobCtx context.Context) error {
			select {
			case <-started:
			default:
				close(started)
			}

			<-jobCtx.Done()
			close(sawCancel)
			return jobCtx.Err()
		})
	}()

	select {
	case <-started:
	case <-time.After(7 * time.Second):
		t.Fatal("timed out waiting for job to start")
	}

	cancel()

	select {
	case <-sawCancel:
	case <-time.After(2 * time.Second):
		t.Fatal("job did not observe cancellation")
	}

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for scheduler to stop")
	}
}

func TestRunLeavesNoExtraGoroutinesAfterExit(t *testing.T) {
	baseline := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := newTestScheduler()
	done := make(chan error, 1)

	go func() {
		done <- scheduler.Run(ctx, func(context.Context) error { return nil })
	}()

	time.Sleep(25 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for scheduler to stop")
	}

	after := runtime.NumGoroutine()
	if after > baseline+5 {
		t.Fatalf("possible goroutine leak: baseline=%d after=%d", baseline, after)
	}
}
