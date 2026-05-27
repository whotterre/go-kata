package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"
)

func TestFetch_ContextCancelsHTTPRequests(t *testing.T) {
	serverStarted := make(chan struct{})
	serverCanceled := make(chan struct{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(serverStarted)
		select {
		case <-r.Context().Done():
			close(serverCanceled)
			return
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer ts.Close()

	fetcher := NewDataFetcher()

	ctx, cancel := context.WithCancel(context.Background())
	resCh := fetcher.Fetch(ctx, []string{ts.URL})

	<-serverStarted

	cancel()

	select {
	case <-serverCanceled:
	case <-time.After(2 * time.Second):
		t.Fatal("server handler did not observe request context cancellation")
	}

	for range resCh {
	}
}

func TestFetch_PartialResultsAndClose(t *testing.T) {
	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fast"))
	}))
	defer fast.Close()

	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(500 * time.Millisecond):
			_, _ = w.Write([]byte("slow"))
		}
	}))
	defer slow.Close()

	fetcher := NewDataFetcher()

	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	resCh := fetcher.Fetch(ctx, []string{fast.URL, slow.URL})

	gotAny := false
	select {
	case r, ok := <-resCh:
		if !ok {
			t.Fatal("results channel closed unexpectedly")
		}
		if r.Error != nil {
			t.Fatalf("unexpected error: %v", r.Error)
		}
		gotAny = true
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for first result")
	}

	if !gotAny {
		t.Fatal("did not receive any result before cancel")
	}

	cancel()

	// draining should finish quickly after cancellation
	done := make(chan struct{})
	go func() {
		for range resCh {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("results channel did not close after cancel")
	}

	after := runtime.NumGoroutine()
	if after > before+10 {
		t.Fatalf("possible goroutine leak: before=%d after=%d", before, after)
	}
}

func TestForgottenSender_NoGoroutineLeak(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer ts.Close()

	urls := make([]string, 100)
	for i := range urls {
		urls[i] = ts.URL
	}

	fetcher := NewDataFetcher()

	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	resCh := fetcher.Fetch(ctx, urls)

	select {
	case <-resCh:
	case <-time.After(500 * time.Millisecond):
	}
	cancel()

	deadline := time.After(3 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			t.Fatalf("goroutines did not settle: before=%d after=%d", before, runtime.NumGoroutine())
		case <-ticker.C:
			if runtime.NumGoroutine() <= before+20 {
				return
			}
		}
	}
}

func TestFetch_CancelBeforeReceive(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
			_, _ = w.Write([]byte("ok"))
		case <-r.Context().Done():
			return
		}
	}))
	defer ts.Close()

	fetcher := NewDataFetcher()

	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	resCh := fetcher.Fetch(ctx, []string{ts.URL, ts.URL, ts.URL})
	cancel()

	done := make(chan struct{})
	go func() {
		for range resCh {
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("results channel did not close after immediate cancel")
	}

	after := runtime.NumGoroutine()
	if after > before+20 {
		t.Fatalf("possible goroutine leak on immediate cancel: before=%d after=%d", before, after)
	}
}
