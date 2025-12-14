package proxy

import (
	"sync"
	"time"
)

type BackendMetrics struct {
	Requests       int64 `json:"requests"`
	Failures       int64 `json:"failures"`
	TotalLatencyMs int64 `json:"total_latency_ms"`
}

type Metrics struct {
	mu            sync.Mutex
	TotalRequests int64                      `json:"total_requests"`
	TotalRetries  int64                      `json:"total_retries"`
	PerBackend    map[string]*BackendMetrics `json:"per_backend"`
}

func NewMetrics() *Metrics {
	return &Metrics{
		PerBackend: make(map[string]*BackendMetrics),
	}
}

func (m *Metrics) ensure(url string) *BackendMetrics {
	if m.PerBackend[url] == nil {
		m.PerBackend[url] = &BackendMetrics{}
	}
	return m.PerBackend[url]
}

func (m *Metrics) IncRequest(backendURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
	m.ensure(backendURL).Requests++
}

func (m *Metrics) IncFailure(backendURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure(backendURL).Failures++
}

func (m *Metrics) IncRetry() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRetries++
}

func (m *Metrics) AddLatency(backendURL string, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure(backendURL).TotalLatencyMs += d.Milliseconds()
}

func (m *Metrics) Snapshot() *Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	copy := &Metrics{
		TotalRequests: m.TotalRequests,
		TotalRetries:  m.TotalRetries,
		PerBackend:    make(map[string]*BackendMetrics),
	}

	for k, v := range m.PerBackend {
		copy.PerBackend[k] = &BackendMetrics{
			Requests:       v.Requests,
			Failures:       v.Failures,
			TotalLatencyMs: v.TotalLatencyMs,
		}
	}
	return copy
}
