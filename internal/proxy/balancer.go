package proxy

import (
	"sync"
	"sync/atomic"
)

type Backend struct {
	URL         string
	activeConns int64
}

func (b *Backend) Inc() {
	atomic.AddInt64(&b.activeConns, 1)
}

func (b *Backend) Dec() {
	atomic.AddInt64(&b.activeConns, -1)
}

func (b *Backend) Active() int64 {
	return atomic.LoadInt64(&b.activeConns)
}

type Balancer interface {
	Next() *Backend
}

// ─── Round Robin ──────────────────────────────────────────────

type RoundRobinBalancer struct {
	backends []*Backend
	idx      int
	mu       sync.Mutex
}

func NewRoundRobin(backends []string) *RoundRobinBalancer {
	bs := make([]*Backend, 0, len(backends))
	for _, u := range backends {
		bs = append(bs, &Backend{URL: u})
	}
	return &RoundRobinBalancer{backends: bs}
}

func (b *RoundRobinBalancer) Next() *Backend {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := len(b.backends)
	if n == 0 {
		return nil
	}
	be := b.backends[b.idx]
	b.idx = (b.idx + 1) % n
	return be
}

// ─── Least Connections ────────────────────────────────────────

type LeastConnBalancer struct {
	backends []*Backend
	mu       sync.Mutex
}

func NewLeastConn(backends []string) *LeastConnBalancer {
	bs := make([]*Backend, 0, len(backends))
	for _, u := range backends {
		bs = append(bs, &Backend{URL: u})
	}
	return &LeastConnBalancer{backends: bs}
}

func (b *LeastConnBalancer) Next() *Backend {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.backends) == 0 {
		return nil
	}
	// pick backend with lowest activeConns
	var chosen *Backend
	for i, be := range b.backends {
		if i == 0 || be.Active() < chosen.Active() {
			chosen = be
		}
	}
	return chosen
}
