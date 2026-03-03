package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Server struct {
	httpServer http.Server
	dbConn     net.Conn
	workerChan chan struct{}
	wg         sync.WaitGroup
}

func NewServer(dbConn net.Conn) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	return &Server{
		httpServer: http.Server{
			Addr:         ":8080",
			Handler:      http.TimeoutHandler(mux, 8*time.Second, "timeout reached"),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		dbConn:     dbConn,
		workerChan: make(chan struct{}, 100),
	}
}

func (s *Server) warmCache(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	// clean up ticker
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// warm cache logic
			log.Println("Warming cache....")
		case <-ctx.Done():
			log.Println("Cache warmer stopping....")
			return
		}
	}
}

func (s *Server) Start(ctx context.Context) error {
	// create a fixed number of goroutines
	n := 10 
	for range n {
		s.wg.Add(1)
		go s.worker(&s.wg)
	}
	go s.warmCache(ctx)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	// stop accepting requests
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP shutdown: %w", err)
	}
	// close the worker channel to signal there are no new
	// jobs to process
	close(s.workerChan)
	// wait till all workers finish their current jobs
	waitDone := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(waitDone)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	case <-waitDone:
		// Everything finished cleanly within the deadline
		log.Println("All workers finished cleanly")
	}
	// close db
	log.Println("Closing database connection...")
	return s.dbConn.Close()
}

func (s *Server) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	// listens for jobs indefinitely
	for range s.workerChan {
		// process jobs
		log.Println("Processing jobs...")
	}
}

func main() {
	// initialize net.Conn via net.Pipe()
	dbConn, _ := net.Pipe()
	// initialize a new instance of the server
	server := NewServer(dbConn)
	// create a context for the start method
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Println("Starting server")
	// start the server in it's own goroutine
	if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}

	// add listeners for SIGTERM and SIGINT
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// block till a signal is recieved
	<-sigChan
	fmt.Println("\nShutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		fmt.Printf("Shutdown failed: %v\n", err)
	}
}
