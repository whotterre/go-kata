package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"runtime"
	"testing"
	"time"
)

// TestGracefulShutdown ensures the server stops workers and closes DB within deadline.
func TestGracefulShutdown(t *testing.T) {
	// Setup
	dbConn, _ := net.Pipe()
	server := NewServer(dbConn)
	
	initialGoroutines := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start Server
	go func() {
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server started with error: %v", err)
		}
	}()

	// Give it a moment to spin up goroutines (workers + warmer)
	time.Sleep(100 * time.Millisecond)
	
	if runtime.NumGoroutine() <= initialGoroutines {
		log.Println(initialGoroutines)
		t.Error("Expected worker goroutines to be spawned, but count did not increase")
	}

	// Trigger Shutdown
	cancel() // Simulate signal by canceling parent context
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	err := server.Stop(shutdownCtx)
	if err != nil {
		t.Errorf("Graceful shutdown failed: %v", err)
	}

	// 4. Verify Leak (Self-Correction Test #2)
	// Give Go scheduler a tiny window to cleanup
	time.Sleep(100 * time.Millisecond)
	finalGoroutines := runtime.NumGoroutine()
	
	if finalGoroutines > initialGoroutines + 2 { // for test runner overhead 
		t.Errorf("Goroutine leak detected: started with %d, ended with %d", initialGoroutines, finalGoroutines)
	}
}

// TestShutdownTimeout ensures the server forces a close when workers take too long.
func TestShutdownTimeout(t *testing.T) {
	dbConn, _ := net.Pipe()
	server := NewServer(dbConn)
	
	// Start server
	ctx := t.Context()
	go server.Start(ctx)

	// Add a "stuck" job to the worker pool
	server.wg.Add(1) 
	go func() {
		defer server.wg.Done()
		time.Sleep(5 * time.Second) // Simulate a very slow task
	}()

	// Try to shut down with a very short timeout
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer shortCancel()

	err := server.Stop(shortCtx)
	if err == nil {
		t.Error("Expected timeout error during shutdown, but got nil")
	}
}