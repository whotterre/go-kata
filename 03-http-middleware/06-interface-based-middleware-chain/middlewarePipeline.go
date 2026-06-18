package main

import (
	"context"
	"log" // replace with slog in a later impl (all hail the odius Procrastinator...inator)
	"sync"
	"time"
)

type Event struct {
	ID   string
	Type string
}


// do we persist collected metrics?
type MetricsCollector interface {
    CountEvent(stage string)
    MeasureLatency(stage string, duration time.Duration)
}

type metricsCollector struct {
	logger *log.Logger
	mu sync.Mutex
	counts map[string]int
}

func (c *metricsCollector) CountEvent(stage string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts[stage]++ 

	c.logger.Printf("[Metrics] Stage %s processed an event. Total: %d", stage, c.counts[stage])
}

func (c *metricsCollector) MeasureLatency(stage string, duration time.Duration) {
	c.logger.Printf("Stage: %s | Latency: %v", stage, duration)
}

func WithMetricsCollector() MetricsCollector {
	return &metricsCollector{}
}

// Validation, enrichment, filtering, and finally storage.
// track metrics at each stage
// how do we achieve recoverability?

type Processor interface {
	Process(context.Context, Event) ([]Event, error)
}

// Validation stage 
type ValidationProcessor struct {
	next Processor
}

type ValidationOption func(*ValidationProcessor)

func NewValidatorProcessor(opts ...ValidationOption) Processor {
	processor := &ValidationProcessor{}

	for _, opt := range opts {
		opt(processor)
	}

	return processor
}

func (p *ValidationProcessor) Process(context.Context, Event) ([]Event, error) {
	return []Event{}, nil
}

// Enrichment stage

type EnrichmentOption func(*EnrichmentProcessor)

type EnrichmentProcessor struct{
	next Processor
}

func NewEnrichmentProcessor(next Processor, opts ...EnrichmentOption) Processor {
    processor := &EnrichmentProcessor{
        next: next, 
    }

    for _, opt := range opts {
        opt(processor)
    }

    return processor
}

func (p *EnrichmentProcessor) Process(context.Context, Event) ([]Event, error) {
	return []Event{}, nil
}


// Filtering stage 
type FilteringProcessor struct {
	next Processor
}

type FilteringOption func(*FilteringProcessor)

func (p *FilteringProcessor) Process(context.Context, Event) ([]Event, error) {
	return []Event{}, nil
}

func NewFilteringProcessor(opts ...FilteringOption) Processor {
	processor := &FilteringProcessor{}

	for _, opt := range opts {
		opt(processor)
	}

	return processor
}

// Storage stage 
type StorageOption func(*StorageProcessor)

type StorageProcessor struct {}

func (p *StorageProcessor) Process(context.Context, Event) ([]Event, error) {
	return []Event{}, nil
}

func NewStorageProcessor(opts ...StorageOption) Processor {
	processor := &StorageProcessor{}

	for _, opt := range opts {
		opt(processor)
	}

	return processor
}


func main() {
	println("Docete me et ergo ipse yada yada...")
}
