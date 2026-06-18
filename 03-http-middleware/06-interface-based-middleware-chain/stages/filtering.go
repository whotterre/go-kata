package stages

import (
	"context"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
)

// Filtering stage
type FilteringProcessor struct {
	next common.Processor
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
