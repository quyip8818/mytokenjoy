package middleware

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/http/response"
)

// RequireSaaS returns 404 for requests when not in SaaS mode.
func RequireSaaS(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.SupportSaas {
				response.Error(w, http.StatusNotFound, "Not found")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireLocal returns 404 for requests when in SaaS mode.
func RequireLocal(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.SupportSaas {
				response.Error(w, http.StatusNotFound, "Not found")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
