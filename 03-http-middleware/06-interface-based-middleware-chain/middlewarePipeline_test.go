package main

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/metrics"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/stages"
)

type AmplifyingProcessor struct {
	next common.Processor
}

func (a *AmplifyingProcessor) Process(ctx context.Context, ev common.Event) ([]common.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ev1 := ev
	ev2 := common.Event{
		ID:      ev.ID + "_copy",
		Type:    ev.Type,
		Message: ev.Message,
	}

	// Pass both independently down the stream contract
	res1, err := a.next.Process(ctx, ev1)
	if err != nil {
		return nil, err
	}

	res2, err := a.next.Process(ctx, ev2)
	if err != nil {
		return nil, err
	}

	return append(res1, res2...), nil
}

type CustomFilterProcessor struct {
	next common.Processor
}

func (f *CustomFilterProcessor) Process(ctx context.Context, ev common.Event) ([]common.Event, error) {
	if strings.Count(ev.ID, "_copy") > 2 {
		return []common.Event{}, nil 
	}
	return f.next.Process(ctx, ev)
}

type TerminalStub struct{}

func (t *TerminalStub) Process(ctx context.Context, ev common.Event) ([]common.Event, error) {
	return []common.Event{ev}, nil
}



func TestIfInfiniteLoop(t *testing.T) {
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	
	stub := &TerminalStub{}
	filter := &CustomFilterProcessor{next: stub}
	amplifier := &AmplifyingProcessor{next: filter}

	ctx := context.Background()
	event := common.Event{
		ID:      "root_evt",
		Type:    "transaction",
		Message: "initial payload",
	}

	finalEvents, err := amplifier.Process(ctx, event)
	if err != nil {
		t.Fatalf("Pipeline blew up under event amplification: %v", err)
	}

	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)


	expectedCount := 4
	if len(finalEvents) != expectedCount {
		t.Errorf("Fail Condition Met: Event multiplication ran uncontrolled. Expected %d events, got %d", expectedCount, len(finalEvents))
	}

	heapGrowth := memAfter.Alloc - memBefore.Alloc
	maxAllowedGrowth := uint64(5 * 1024 * 1024)
	if heapGrowth > maxAllowedGrowth {
		t.Errorf("Fail Condition Met: Memory usage grew exponentially. Heap growth: %d bytes", heapGrowth)
	}
}

func TestCancelsAfterTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Trigger manual context cancellation mid-flight after 50ms
	time.AfterFunc(50*time.Millisecond, func() {
		cancel()
	})

	myMetrics := metrics.NewMetricsCollector()
	myStorer := stages.NewFileStorer()

	// Orchestrate the real processing chain backward
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
		ID:      "test_abort_id",
		Type:    "audit",
		Message: "sensitive data payload",
	}

	_, err := pipeline.Process(ctx, event)
	if err == nil {
		t.Fatal("Fail Condition Met: Pipeline continued processing completely after context cancellation")
	}
	
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Expected context cancellation error, but caught an alternate failure structure: %v", err)
	}
}