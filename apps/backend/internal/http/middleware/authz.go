package middleware

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

func RequireAnyPermission(required ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionCtx, ok := SessionFromContext(r.Context())
			if !ok {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}
			if !permission.HasAny(sessionCtx.Permissions, required...) {
				httputil.WriteStatus(w, http.StatusForbidden, httputil.MsgForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
