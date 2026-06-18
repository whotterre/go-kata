package stages

import (
	"context"
	"time"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/metrics"
)

// Filtering stage
type FilteringProcessor struct {
	next    common.Processor
	metrics metrics.MetricsCollector
}

type FilteringOption func(*FilteringProcessor)

func (p *FilteringProcessor) Process(ctx context.Context, ev common.Event) ([]common.Event, error) {
	start := time.Now()
	defer func() {
		if p.metrics != nil {
			p.metrics.MeasureLatency("filtering", time.Since(start))
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if ev.Type == "debug_log" {
		p.metrics.CountEvent("drop_event@")
		return []common.Event{}, nil
	}


	return p.next.Process(ctx, ev)
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