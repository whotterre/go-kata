package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

type FanOutClient struct {
	httpClient *http.Client
	limiter    *rate.Limiter
	sem        *semaphore.Weighted
	logger     *slog.Logger
}

func NewFanOutClient() *FanOutClient {
	customTransport := &http.Transport{
		MaxIdleConnsPerHost: 8,
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: customTransport,
	}
	return &FanOutClient{
		httpClient: client,
		limiter:    rate.NewLimiter(rate.Limit(10), 20),
		sem:        semaphore.NewWeighted(8), // max 8 connections for the connection pool
		logger:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
}

func (c *FanOutClient) DoRequest(ctx context.Context, id int) ([]byte, error) {
	url := "https://api.thecatapi.com/v1/images/search?size=med&mime_types=jpg&format=json&has_breeds=true&order=RANDOM&page=0"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func (c *FanOutClient) FetchAll(ctx context.Context, userIDs []int) (map[int][]byte, error) {
	g, ctx := errgroup.WithContext(ctx)

	results := make(map[int][]byte)
	var mu sync.Mutex

	for _, id := range userIDs {
		id := id

		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}

		// Wait for Semaphore (Max 8 in-flight)
		if err := c.sem.Acquire(ctx, 1); err != nil {
			return nil, err
		}

		g.Go(func() error {
			defer c.sem.Release(1)
			start := time.Now()

			// make the request
			data, err := c.DoRequest(ctx, id)
			if err != nil {
				return err
			}
			// write to result map
			mu.Lock()
			results[id] = data
			mu.Unlock()

			slog.Info("fetched fact", "userID", id, "latency", time.Since(start))
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func main() {
	context := context.Background()
	f := NewFanOutClient()
	users := []int{1, 2, 3, 4, 5, 6, 7}
	res, err := f.FetchAll(context, users)
	if err != nil {
		println(err)
	}
	fmt.Println(res)
}
