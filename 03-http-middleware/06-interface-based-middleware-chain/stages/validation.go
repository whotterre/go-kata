package stages

import (
	"context"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
)

// Validation stage (First stage)
type ValidationProcessor struct {
	next common.Processor
}

type ValidationOption func(*ValidationProcessor)

func NewValidatorProcessor(opts ...ValidationOption) common.Processor {
	processor := &ValidationProcessor{}

	for _, opt := range opts {
		opt(processor)
	}

	return processor
}

func (p *ValidationProcessor) Process(context.Context, common.Event) ([]common.Event, error) {
	return []common.Event{}, nil
}
