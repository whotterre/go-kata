package stages

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"time"
	"log"

	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/common"
	"github.com/medunes/go-kata/03-http-middleware/06-interface-based-middleware-chain/metrics"
)

// Storage stage
type StorageOption func(*StorageProcessor)

type StorageProcessor struct {
	metrics metrics.MetricsCollector
	storer Storer 
}

type Storer interface {
	Store(ev common.Event) (string, error)
}



func (p *StorageProcessor) Process(ctx context.Context, ev common.Event) ([]common.Event, error) {
    start := time.Now()
    defer func() {
        if p.metrics != nil {
            p.metrics.MeasureLatency("storage", time.Since(start)) // <-- Fixed label from "filter"
        }
    }()
    if err := ctx.Err(); err != nil {
        return nil, err
    }

    hash, err := p.storer.Store(ev)
	log.Println("Stored file successfully with hash", hash)
    if err != nil {
        if p.metrics != nil { 
            p.metrics.CountEvent("storage_error")
        }
        return nil, err
    }
    
    if p.metrics != nil {
       p.metrics.CountEvent("storage_success")
    }

    return []common.Event{ev}, nil 
}


type fileStorer struct {}

func NewFileStorer() Storer {
	return &fileStorer{}
}

func (s *fileStorer) Store(ev common.Event) (string, error){
	file, err := os.Open(ev.ID)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(ev)
	if err != nil {
		return "", errors.New("Error in marshalling event")
	}
	_, err = file.Write(data)
	if err != nil {
		return "", errors.New("Failed to write event data to file")
	}
	
	// hash the fields 
	payload := ev.ID + ev.Message + ev.Type
	hash := string(sha256.New().Sum([]byte(payload)))
	// store the kini
	return hash, nil
}

func NewStorageProcessor(storer Storer, opts ...StorageOption) common.Processor {
	processor := &StorageProcessor{
		storer: storer,
	}

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