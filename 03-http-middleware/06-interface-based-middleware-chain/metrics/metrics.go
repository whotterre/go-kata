package metrics

import (
	"log" // replace with slog in a later impl (all hail the odius Procrastinator...inator)
	"os"
	"sync"
	"time"
)

// do we persist collected metrics?
type MetricsCollector interface {
    CountEvent(stage string)
    MeasureLatency(stage string, duration time.Duration)
}

func NewMetricsCollector() MetricsCollector {
	return &metricsCollector{
		logger: log.New(os.Stdout, "[METRICS] ", log.LstdFlags),
		counts: make(map[string]int),
	}
}

type metricsCollector struct {
	logger *log.Logger
	mu sync.Mutex
	counts map[string]int
}

func (c *metricsCollector) CountEvent(stage string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts[stage]++ 

	c.logger.Printf("[Metrics] Stage %s processed an event. Total: %d", stage, c.counts[stage])
}

func (c *metricsCollector) MeasureLatency(stage string, duration time.Duration) {
	c.logger.Printf("Stage: %s | Latency: %v", stage, duration)
}

func WithMetricsCollector() MetricsCollector {
	return &metricsCollector{}
}