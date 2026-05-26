package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

type Result struct {
	URL string 
	Body []byte
	Error error
}


type DataFetcher struct {
	client *http.Client	
}

func NewDataFetcher() *DataFetcher {
	return &DataFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (f *DataFetcher) Fetch(ctx context.Context, urls []string) <-chan Result {
	resChan := make(chan Result)

	g, ctx := errgroup.WithContext(ctx)
	
	for _, url := range urls {
		url := url 
		g.Go(func() error {
		  resp, err := f.client.Get(url)
		  if err != nil {
			resChan <- Result{URL: url, Error: err}
			return nil 
		  }
		  defer resp.Body.Close()
		  
		  body, err := io.ReadAll(resp.Body)
		  if err != nil {
				resChan <- Result{URL: url, Error: err}
				return nil
		  }
		  resChan <- Result{
			URL: url,
			Body: body,
			Error: err,
		  }
		  return nil
		})
	}
	
	go func () {
		_ = g.Wait() // wait for all goroutines to exec
		close(resChan) // close chan when complete
	}()
	return resChan
}



func main(){
	urls := []string{
		"https://google.com",
		"https://kibo.com",
		"https://neeto.com",
	}
	
	// two stages: fetch urls and stream downstream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// create a new fetcher 
	fetcher := NewDataFetcher()

	results := fetcher.Fetch(ctx, urls)

	for result := range results {
		if result.Error != nil {
			fmt.Printf("Error fetching %s: %v\n", result.URL, result.Error)
			continue
		}
		fmt.Printf("Fetched %s: %d bytes\n", result.URL, len(result.Body))
	}

}
