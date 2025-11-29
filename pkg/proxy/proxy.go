package proxy

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Proxy struct {
	cfg     *Config
	router  *Router
	metrics *Metrics
	server  *http.Server
}

func New(cfg *Config) *Proxy {
	router := NewRouter(cfg)
	metrics := NewMetrics()

	handler := &coreHandler{
		router:  router,
		metrics: metrics,
		timeout: cfg.Timeout(),
	}

	handlerWithChain := Chain(handler) // middleware pluggable here

	return &Proxy{
		cfg:     cfg,
		router:  router,
		metrics: metrics,
		server:  &http.Server{Addr: cfg.ListenAddr, Handler: handlerWithChain},
	}
}

func (p *Proxy) Start() error {
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		log.Println("shutting down gracefully...")
		p.server.Close()
	}()

	log.Println("proxy running on", p.cfg.ListenAddr)
	return p.server.ListenAndServe()
}
