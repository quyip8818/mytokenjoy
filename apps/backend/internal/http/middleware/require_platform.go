package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/identity/httpx"
)

// RequirePlatformAdmin rejects requests unless the session belongs to the super
// company (tokenJoyCompanyID) AND carries the "platform:manage" permission.
func RequirePlatformAdmin(tokenJoyCompanyID uuid.UUID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := httpx.SessionFromContext(r.Context())
			if !ok || session.CompanyID != tokenJoyCompanyID {
				httputil.WriteStatus(w, http.StatusForbidden, httputil.MsgForbidden)
				return
			}
			if !authz.HasAny(session.Permissions, "platform:manage") {
				httputil.WriteStatus(w, http.StatusForbidden, httputil.MsgForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
