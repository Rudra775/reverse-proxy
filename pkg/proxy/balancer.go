package proxy

import (
	"errors"
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL         string
	activeConns int64
	healthy     int32 // 1 = healthy, 0 = unhealthy
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

func (b *Backend) IsHealthy() bool {
	// default to healthy if never set
	if atomic.LoadInt32(&b.healthy) == 0 {
		return true
	}
	return atomic.LoadInt32(&b.healthy) == 1
}

func (b *Backend) SetHealthy(v bool) {
	if v {
		atomic.StoreInt32(&b.healthy, 1)
	} else {
		atomic.StoreInt32(&b.healthy, -1)
	}
}

// ───────────────────────────────────────────
// Balancer Interface

type Balancer interface {
	Next() (*Backend, error)
}

// ───────────────────────────────────────────
// Round Robin (skips unhealthy backends)

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

	n := len(rr.backends)
	if n == 0 {
		return nil, errors.New("no backends")
	}

	// Try all backends at most once
	for i := 0; i < n; i++ {
		idx := (rr.index + i) % n
		be := rr.backends[idx]
		if be.IsHealthy() {
			rr.index = (idx + 1) % n
			return be, nil
		}
	}
	return nil, errors.New("no healthy backends")
}
