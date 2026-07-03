package middleware

import (
	"net/http"

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
			if p.Cfg.SyncTriggerAPIKey != "" && r.Header.Get(SyncTriggerAPIKeyHeader) == p.Cfg.SyncTriggerAPIKey {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}
