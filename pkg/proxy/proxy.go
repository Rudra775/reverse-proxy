package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Rudra775/reverse-proxy/internal/server"
)

// Proxy is the main reverse proxy instance.
// It wires router, handlers, middleware, metrics and server lifecycle.
type Proxy struct {
	cfg         *Config
	router      *Router
	metrics     *Metrics
	cache       *LRUCache
	mux         *http.ServeMux
	middlewares []Middleware
}

// Use registers a middleware to be applied globally.
func (p *Proxy) Use(m Middleware) {
	p.middlewares = append(p.middlewares, m)
}

// New constructs a new Proxy instance.
// Heavy logic lives in handlers; this only wires components together.
func New(cfg *Config) *Proxy {
	router := NewRouter(cfg)
	metrics := NewMetrics()
	cache := NewLRUCache(cfg.CacheBytes())

	core := &coreHandler{
		router:   router,
		metrics:  metrics,
		cache:    cache,
		timeout:  cfg.Timeout(),
		maxRetry: cfg.MaxRetries,
	}

	mux := http.NewServeMux()
	mux.Handle("/", core)
	mux.HandleFunc("/metrics", metricsHandler(metrics))

	p := &Proxy{
		cfg:     cfg,
		router:  router,
		metrics: metrics,
		cache:   cache,
		mux:     mux,
	}

	// Start active health checks if configured
	if interval := cfg.HealthCheckInterval(); interval > 0 {
		go p.startHealthChecks(interval)
	}

	return p
}

// Start launches the HTTP server with middleware chain applied.
func (p *Proxy) Start() error {
	var handler http.Handler = p.mux

	// Apply middleware chain (outermost first)
	handler = Chain(handler, p.middlewares...)

	srv := server.New(p.cfg.ListenAddr, handler)
	return srv.Start()
}

// Shutdown allows programmatic shutdown when embedding as a library.
func (p *Proxy) Shutdown(ctx context.Context) error {
	// internal/server already handles signal-based shutdown
	// this hook is useful for embedding use-cases
	_ = ctx
	return nil
}

// ─────────────────────────────────────────────
// Metrics handler

func metricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := metrics.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snapshot)
	}
}

// ─────────────────────────────────────────────
// Active health checking (HEAD requests)

func (p *Proxy) startHealthChecks(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	client := &http.Client{Timeout: 2 * time.Second}

	for range ticker.C {
		for _, rt := range p.router.routes {
			rr, ok := rt.balancer.(*RoundRobin)
			if !ok {
				continue
			}

			for _, be := range rr.backends {
				go func(b *Backend) {
					resp, err := client.Head(b.URL)
					if err != nil || resp.StatusCode >= 500 {
						b.SetHealthy(false)
						return
					}
					b.SetHealthy(true)
				}(be)
			}
		}
	}
}
