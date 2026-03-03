package main

import (
	"net/http"
	"context"
	"os/signal"
	"sync"
)

type Job struct {

}

type Server struct {
	httpServer http.Server
	dbConn net.Conn
	workerChan chan Job
	wg sync.WaitGroup
}

func NewServer(dbConn net.Conn) *Server {
	mux := http.NewServeMux()
	return &Server{
		httpServer: http.Server{
			Address: ":8080",
			Handler: mux
		},
		dbConn: dbConn,
		workerChan = make(chan Job, 100),
	}
}

var Options () *Server   


func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	// create a fixed number of goroutines
	n := 10
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(ctx, s.jobChan)
	}
}

func (s *Server) Stop(ctx context.Context) error{

}

func worker(ctx context.Context, job chan Job, wg *sync.WaitGroup){
	// listens for jobs
	currJob <- job 
	fmt.Println()
}

func main(){
	// initialize a new instance of the server
	server := NewServer(
		
	)
	
}