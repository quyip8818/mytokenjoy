package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain/company"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

const SyncTriggerAPIKeyHeader = "X-Sync-API-Key"

func AllowSyncTrigger(p httpdeps.Protected) func(http.Handler) http.Handler {
	sessionChain := RequireSession(p)
	authzChain := RequireAnyPermission(permission.OrgDatasource)

	return func(next http.Handler) http.Handler {
		protected := sessionChain(authzChain(next))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerKey := r.Header.Get(SyncTriggerAPIKeyHeader)
			if p.Cfg.SyncTriggerAPIKey != "" && headerKey != "" &&
				subtle.ConstantTimeCompare([]byte(headerKey), []byte(p.Cfg.SyncTriggerAPIKey)) == 1 {
				// Bind request to default company context so downstream handlers have a valid tenant.
				ctx := company.WithDefaultCompany(r.Context(), p.Cfg.DefaultCompanyID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}
