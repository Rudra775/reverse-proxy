package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server wraps the HTTP server with graceful shutdown management.
type Server struct {
	httpServer *http.Server
}

// New constructs a new server instance with the given addr + handler.
func New(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

// Start runs the server and listens for shutdown signals.
func (s *Server) Start() error {
	idleConnsClosed := make(chan struct{})

	// Listen for SIGINT / SIGTERM for graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		log.Printf("[SERVER] Shutdown signal received: %v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("[SERVER] Error during shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("[SERVER] Listening on %s", s.httpServer.Addr)

	// Start HTTP server (blocking call)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	log.Println("[SERVER] Shutdown complete")
	return nil
}

// Stop triggers graceful shutdown programmatically.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
