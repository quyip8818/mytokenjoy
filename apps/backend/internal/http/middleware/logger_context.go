package middleware

import (
	"context"
	"log/slog"
	"net/http"
)

type loggerKey struct{}

// LoggerContext injects a per-request logger (with request_id) into context.
func LoggerContext(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := RequestIDFromContext(r.Context())
			logger := base.With("request_id", reqID)
			ctx := context.WithValue(r.Context(), loggerKey{}, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggerFromContext retrieves the per-request logger. Falls back to slog.Default().
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
