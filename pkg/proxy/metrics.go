package proxy

import "sync"

type Metrics struct {
	mu            sync.Mutex
	TotalRequests int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) Inc() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
}
