package stages

import (
	"strings"
	"context"
	"math/rand/v2"
	"time"

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


func (p *EnrichmentProcessor) Process(ctx context.Context, ev common.Event) ([]common.Event, error) {
    start := time.Now()
	defer func() {
		if p.metrics != nil {
			p.metrics.MeasureLatency("enrichment", time.Since(start))
		}
	}()
    if err := ctx.Err(); err != nil {
		return nil, err
	}
    
    // Enrich here 
    ev.Message = genRandomString(8)
    if p.metrics != nil {
       p.metrics.CountEvent("enrichment") 
    }
	return p.next.Process(ctx, ev)
}

func WithEnrichmentMetricsCollector(collector metrics.MetricsCollector) EnrichmentOption {
	return func(p *EnrichmentProcessor) {
		p.metrics = collector
	}
}

func genRandomString(length int) string {
    var res strings.Builder
    charSet := "abcdefghijklmnopqrstuvwxyz"
    for range length {
        chosen := rand.IntN(len(charSet)) 
        res.WriteString(string(charSet[chosen]))
    }
    return res.String()
}