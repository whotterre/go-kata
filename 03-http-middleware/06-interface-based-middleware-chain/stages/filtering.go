package stages

import (
	"context"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/metrics"
)

// Filtering stage
type FilteringProcessor struct {
	next common.Processor
	metrics metrics.MetricsCollector
}

type FilteringOption func(*FilteringProcessor)

func (p *FilteringProcessor) Process(context.Context, common.Event) ([]common.Event, error) {
	return []common.Event{}, nil
}

func NewFilteringProcessor(opts ...FilteringOption) common.Processor {
	processor := &FilteringProcessor{}

	for _, opt := range opts {
		opt(processor)
	}

	return processor
}


func WithFilteringMetricsCollector(collector metrics.MetricsCollector) FilteringOption {
	return func(p *FilteringProcessor) {
		p.metrics = collector
	}
}