package proxy

import (
	"errors"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL         string
	activeConns int64
}

func (b *Backend) Active() int64 {
	return atomic.LoadInt64(&b.activeConns)
}
func (b *Backend) Inc() {
	atomic.AddInt64(&b.activeConns, 1)
}
func (b *Backend) Dec() {
	atomic.AddInt64(&b.activeConns, -1)
}

// ───────────────────────────────────────────
// Balancer Interface

type Balancer interface {
	Next() (*Backend, error)
}

// ───────────────────────────────────────────
// Round Robin

type RoundRobin struct {
	mu       sync.Mutex
	backends []*Backend
	index    int
}

func NewRoundRobin(backends []string) *RoundRobin {
	bs := make([]*Backend, len(backends))
	for i, u := range backends {
		bs[i] = &Backend{URL: u}
	}
	return &RoundRobin{backends: bs}
}

func (rr *RoundRobin) Next() (*Backend, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.backends) == 0 {
		return nil, errors.New("no backends")
	}

	be := rr.backends[rr.index]
	rr.index = (rr.index + 1) % len(rr.backends)
	return be, nil
}
