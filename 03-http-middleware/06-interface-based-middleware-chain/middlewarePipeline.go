package main

import (
	"context"
	"log"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/metrics"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/stages"
)

// Validation, enrichment, filtering, and finally storage.
// track metrics at each stage
// how do we achieve recoverability?

func main() {
	ctx := context.Background()
	myMetrics := metrics.NewMetricsCollector()
	myStorer := stages.NewFileStorer()

	storageStage := stages.NewStorageProcessor(
		myStorer,
		stages.WithStorageMetricsCollector(myMetrics),
	)

	filteringStage := stages.NewFilteringProcessor(
		storageStage,
		stages.WithFilteringMetricsCollector(myMetrics),
	)

	enrichmentStage := stages.NewEnrichmentProcessor(
		filteringStage,
		stages.WithEnrichmentMetricsCollector(myMetrics),
	)

	pipeline := stages.NewValidatorProcessor(
		enrichmentStage,
		stages.WithValidationMetricsCollector(myMetrics),
	)

	event := common.Event{
		ID:   "1",
		Type: "audit",
	}

	finalEvents, err := pipeline.Process(ctx, event)
	if err != nil {
		log.Fatalf("Pipeline execution failed: %v", err)
	}

	log.Printf("Pipeline completed successfully! Processed %d items.\n", len(finalEvents))
}
