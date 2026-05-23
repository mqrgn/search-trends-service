package metrics

import (
	"sync"
)

type Metrics struct {
	mu                sync.RWMutex
	requestCount      map[string]int64
	requestDuration   map[string]float64
	eventsProcessed   int64
	anomaliesDetected int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		requestCount:    make(map[string]int64),
		requestDuration: make(map[string]float64),
	}
}

func (m *Metrics) RecordRequest(endpoint string, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount[endpoint]++
	m.requestDuration[endpoint] += duration
}

func (m *Metrics) IncrementEventsProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventsProcessed++
}

func (m *Metrics) IncrementAnomaliesDetected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.anomaliesDetected++
}

func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["events_processed"] = m.eventsProcessed
	stats["anomalies_detected"] = m.anomaliesDetected
	stats["request_count"] = m.requestCount
	stats["request_duration"] = m.requestDuration

	return stats
}
