package middleware

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

const SyncTriggerAPIKeyHeader = "X-Sync-API-Key"

func AllowSyncTrigger(cfg config.Config, sessionSvc session.Service) func(http.Handler) http.Handler {
	sessionChain := RequireSession(cfg, sessionSvc)
	authzChain := RequireAnyPermission(permission.OrgDatasource)

	return func(next http.Handler) http.Handler {
		protected := sessionChain(authzChain(next))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.SyncTriggerAPIKey != "" && r.Header.Get(SyncTriggerAPIKeyHeader) == cfg.SyncTriggerAPIKey {
				next.ServeHTTP(w, r)
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}
