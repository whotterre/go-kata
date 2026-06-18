package stages

import (
	"context"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
)

// Enrichment stage

type EnrichmentOption func(*EnrichmentProcessor)

type EnrichmentProcessor struct{
	next common.Processor
}

func NewEnrichmentProcessor(next common.Processor, opts ...EnrichmentOption) common.Processor {
    processor := &EnrichmentProcessor{
        next: next, 
    }

    for _, opt := range opts {
        opt(processor)
    }

    return processor
}

func (p *EnrichmentProcessor) Process(context.Context, common.Event) ([]common.Event, error) {
	return []common.Event{}, nil
}
