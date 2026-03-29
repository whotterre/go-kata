package main

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
)

type Job func(ctx context.Context) error

type Pool struct {
	NumWorkers       int
	StopOnFirstError bool
}

func NewPool(numWorkers int) *Pool {
	return &Pool{
		NumWorkers: numWorkers,
	}
}

func (r *Pool) Run(parentCtx context.Context, jobs <-chan Job) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	// create a dedicated err channel
	errChan := make(chan error, r.NumWorkers)

	// send err to errChan
	for range r.NumWorkers {
		g.Go(func() error {
			// consume the jobs from the worker
			for job := range jobs {
				// Check if we should even start the next job
				if ctx.Err() != nil {
					return ctx.Err()
				}

				if err := job(ctx); err != nil {
					if r.StopOnFirstError {
						errChan <- err
						return err
					}
					errChan <- err
				}
			}
			return nil
		})
	}
	// monitor goroutine to close error channel and release resources allocated to context's calling fns
	go func() {
		g.Wait()
		close(errChan)
		cancel()
	}()

	// aggregate errors
	var errSlice []error
	for err := range errChan {
		errSlice = append(errSlice, err)
	}
	return errors.Join(errSlice...)
}
