package proxy

import "net/http"

// Middleware defines a standard HTTP middleware
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in reverse order (last added runs first)
func Chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}
