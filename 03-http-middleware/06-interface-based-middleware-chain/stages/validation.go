package stages

import (
	"context"
	"errors"
	"time"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/metrics"
)

// Validation stage (First stage)
type ValidationProcessor struct {
	next    common.Processor
	metrics metrics.MetricsCollector
}

type ValidationOption func(*ValidationProcessor)

func NewValidatorProcessor(opts ...ValidationOption) common.Processor {
	processor := &ValidationProcessor{}

	for _, opt := range opts {
		opt(processor)
	}

	return processor
}

func (p *ValidationProcessor) Process(ctx context.Context, ev common.Event) ([]common.Event, error) {
	start := time.Now()
	defer func() {
		if p.metrics != nil {
			p.metrics.MeasureLatency("validation", time.Since(start))
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if ev.ID == "" {
		return nil, errors.New("Event with invalid event ID passed validation stage")
	}

	var res []common.Event

	res = append(res, ev)

	if p.metrics != nil {
		p.metrics.CountEvent("validation")
	}
	return p.next.Process(ctx, ev)
}

func WithValidationMetricsCollector(collector metrics.MetricsCollector) ValidationOption {
	return func(p *ValidationProcessor) {
		p.metrics = collector
	}
}
