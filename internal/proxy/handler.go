package proxy

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"proxyserver/internal/config"
)

type ProxyHandler struct {
	router    *Router
	transport *http.Transport
	timeout   time.Duration
}

func NewProxyHandler(cfg *config.Config, router *Router) *ProxyHandler {
	tr := &http.Transport{
		Proxy:               nil,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
		MaxIdleConnsPerHost: 10,
	}
	return &ProxyHandler{
		router:    router,
		transport: tr,
		timeout:   time.Duration(cfg.RequestTimeoutMs) * time.Millisecond,
	}
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	be := h.router.MatchBackend(r)
	if be == nil {
		http.Error(w, "no backend available", http.StatusBadGateway)
		return
	}

	be.Inc()
	defer be.Dec()

	// Build backend request
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	req := r.Clone(ctx)
	req.RequestURI = "" // required by http.Client / Transport

	// Rewrite URL for backend
	req.URL.Scheme = "http"
	if r.URL.Scheme != "" {
		req.URL.Scheme = r.URL.Scheme
	}
	req.URL.Host = stripScheme(be.URL)
	// For simple case: be.URL is "http://host:port"
	// We'll parse it better in future phases, but this works with "http://localhost:9001"

	// Add proxy headers
	req.Header.Add("X-Forwarded-For", clientIP(r))
	req.Header.Add("X-Forwarded-Host", r.Host)
	req.Header.Add("X-Forwarded-Proto", r.URL.Scheme)

	// Forward to backend
	resp, err := h.transport.RoundTrip(req)
	if err != nil {
		log.Printf("[ERROR] backend %s error: %v", be.URL, err)
		http.Error(w, "backend error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	// Status code
	w.WriteHeader(resp.StatusCode)
	// Body
	_, _ = io.Copy(w, resp.Body)

	duration := time.Since(start)
	log.Printf("[INFO] %s %s → %s %d (%s)",
		r.Method, r.URL.Path, be.URL, resp.StatusCode, duration.String())
}

func clientIP(r *http.Request) string {
	// Try X-Real-IP / X-Forwarded-For etc. later; for now, extract from RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func stripScheme(u string) string {
	// very small helper; assumes "http://host:port"
	if len(u) >= 7 && u[:7] == "http://" {
		return u[7:]
	}
	if len(u) >= 8 && u[:8] == "https://" {
		return u[8:]
	}
	return u
}
