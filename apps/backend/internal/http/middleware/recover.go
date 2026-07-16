package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/tokenjoy/backend/internal/http/httputil"
)

// Recover catches panics in downstream handlers, logs the error with
// stack trace and request ID, then returns 500.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					reqID := RequestIDFromContext(r.Context())
					stack := string(debug.Stack())
					logger.Error("panic recovered",
						"error", recovered,
						"path", r.URL.Path,
						"method", r.Method,
						"request_id", reqID,
						"stack", stack,
					)
					httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
