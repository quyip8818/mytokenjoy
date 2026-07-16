package middleware

import (
	"context"
	"net/http"
	"time"
)

// RequestTimeout applies a context deadline to each request.
// Only used for /api routes; /v1 streaming routes should NOT use this.
func RequestTimeout(seconds int) func(http.Handler) http.Handler {
	if seconds <= 0 {
		seconds = 30
	}
	timeout := time.Duration(seconds) * time.Second

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
