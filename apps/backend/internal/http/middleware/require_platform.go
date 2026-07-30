package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/identity/httpx"
)

// RequirePlatformAdmin rejects requests unless:
// 1. The deployment is SaaS mode (defense-in-depth: local can never reach platform APIs)
// 2. The session belongs to the super company (tokenJoyCompanyID)
// 3. The session carries "platform:manage" permission
func RequirePlatformAdmin(tokenJoyCompanyID uuid.UUID, supportSaas bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !supportSaas {
				httputil.WriteStatus(w, http.StatusForbidden, httputil.MsgForbidden)
				return
			}
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
