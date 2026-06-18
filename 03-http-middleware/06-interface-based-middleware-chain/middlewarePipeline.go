package main

import (
	"context"
	"time"
)

type Event struct {
	ID   string
	Type string
}
// Validation, enrichment, filtering, and finally storage.
// track metrics at each stage)
// how do we achieve recoverablitity
type Processor interface {
	Process(context.Context, Event) ([]Event, error)
}

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

}
