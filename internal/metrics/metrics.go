package metrics

import (
	"sync"
	"time"
)

type BackendMetrics struct {
	Requests  int64
	Failures  int64
	LatencyMs int64 // sum of latencies, used for avg
}

type Metrics struct {
	mu            sync.Mutex
	TotalRequests int64
	TotalRetries  int64

	PerBackend map[string]*BackendMetrics
}

func New() *Metrics {
	return &Metrics{
		PerBackend: make(map[string]*BackendMetrics),
	}
}

func (m *Metrics) IncRequest(backend string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	bm := m.ensure(backend)
	bm.Requests++
}

func (m *Metrics) IncFailure(backend string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	bm := m.ensure(backend)
	bm.Failures++
}

func (m *Metrics) AddLatency(backend string, dur time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	bm := m.ensure(backend)
	bm.LatencyMs += dur.Milliseconds()
}

func (m *Metrics) IncRetry() {
	m.mu.Lock()
	m.TotalRetries++
	m.mu.Unlock()
}

func (m *Metrics) ensure(backend string) *BackendMetrics {
	if m.PerBackend[backend] == nil {
		m.PerBackend[backend] = &BackendMetrics{}
	}
	return m.PerBackend[backend]
}
