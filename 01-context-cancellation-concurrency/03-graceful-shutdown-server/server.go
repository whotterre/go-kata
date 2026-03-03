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
	n := 10 // TODO: make this configurable
	for range n {
		s.wg.Add(1)
		go worker(ctx, s.workerChan)
	}
	go s.warmCache(ctx)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}

func worker(ctx context.Context, job chan struct{}) {
	// listens for jobs
	// currJob := <-workerChan
	fmt.Println()
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
	go server.Start(ctx)

	// add a listener for SIGTERM
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
