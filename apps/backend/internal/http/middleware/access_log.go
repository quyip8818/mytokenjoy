package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
)

// AccessLog records a structured JSON log entry for each HTTP request.
// It captures method, path, status, latency, request_id, company_id, and client IP.
// Requests to /healthz are skipped.
func AccessLog(logger *slog.Logger, slowThresholdMs int) func(http.Handler) http.Handler {
	if slowThresholdMs <= 0 {
		slowThresholdMs = 5000
	}
	slowThreshold := time.Duration(slowThresholdMs) * time.Millisecond

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip health checks.
			if r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)
			latency := time.Since(start)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"latency_ms", latency.Milliseconds(),
				"request_id", RequestIDFromContext(r.Context()),
				"ip", r.RemoteAddr,
				"bytes", wrapped.bytes,
			}

			if companyID := ctxcompany.ID(r.Context()); companyID > 0 {
				attrs = append(attrs, "company_id", companyID)
			}

			if latency >= slowThreshold {
				attrs = append(attrs, "slow", true)
			}

			logger.LogAttrs(r.Context(), slog.LevelInfo, "http_request", attrsToSlog(attrs)...)
		})
	}
}

func attrsToSlog(pairs []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		key, _ := pairs[i].(string)
		attrs = append(attrs, slog.Any(key, pairs[i+1]))
	}
	return attrs
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// Unwrap supports http.ResponseController and middleware that check for
// wrapped writers (e.g. http.Flusher).
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
