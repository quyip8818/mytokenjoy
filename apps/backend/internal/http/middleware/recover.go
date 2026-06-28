package middleware

import (
	"log/slog"
	"net/http"

	"github.com/tokenjoy/backend/internal/http/httputil"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered", "error", recovered, "path", r.URL.Path)
					httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
