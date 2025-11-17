package proxy

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proxyserver/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return &Server{httpServer: srv}
}

func (s *Server) Start() error {
	// graceful shutdown
	idleConnsClosed := make(chan struct{})

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutdown signal received, draining connections...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("Reverse proxy listening on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	return nil
}
