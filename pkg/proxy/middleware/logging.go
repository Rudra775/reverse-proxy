package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(start)

			reqID, _ := r.Context().Value(RequestIDKey).(string)

			log.Printf(
				`{"request_id":"%s","method":"%s","path":"%s","remote":"%s","latency_ms":%d}`,
				reqID,
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				duration.Milliseconds(),
			)
		})
	}
}
