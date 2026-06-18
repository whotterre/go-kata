package stages

import (
	"context"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/metrics"
)

// Enrichment stage

type EnrichmentOption func(*EnrichmentProcessor)

type EnrichmentProcessor struct{
	next common.Processor
    metrics metrics.MetricsCollector

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

func WithEnrichmentMetricsCollector(collector metrics.MetricsCollector) EnrichmentOption {
	return func(p *EnrichmentProcessor) {
		p.metrics = collector
	}
}