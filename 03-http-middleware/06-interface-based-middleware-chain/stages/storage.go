package stages

import (
	"context"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/metrics"
)

// Storage stage
type StorageOption func(*StorageProcessor)

type StorageProcessor struct {
	metrics metrics.MetricsCollector
}

func (p *StorageProcessor) Process(context.Context, common.Event) ([]common.Event, error) {
	return []common.Event{}, nil
}

func NewStorageProcessor(opts ...StorageOption) common.Processor {
	processor := &StorageProcessor{}

	for _, opt := range opts {
		opt(processor)
	}

	return processor
}

func WithStorageMetricsCollector(collector metrics.MetricsCollector) StorageOption {
	return func(p *StorageProcessor) {
		p.metrics = collector
	}
}