package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = uuid.NewString()
			}

			ctx := context.WithValue(r.Context(), RequestIDKey, id)
			r = r.WithContext(ctx)

			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r)
		})
	}
}
