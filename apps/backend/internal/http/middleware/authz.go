package middleware

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/permission"
)

func RequireAnyPermission(required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionCtx, ok := SessionFromContext(r.Context())
			if !ok {
				response.Error(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			if !permission.HasAny(sessionCtx.Permissions, required...) {
				response.Error(w, http.StatusForbidden, "Forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
