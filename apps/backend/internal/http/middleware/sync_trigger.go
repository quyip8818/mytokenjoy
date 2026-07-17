package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/google/uuid"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

const SyncTriggerAPIKeyHeader = "X-Sync-API-Key"
const CompanyIDHeader = "X-Company-ID"

func AllowSyncTrigger(p httpdeps.Protected, companySvc domaincompany.Service) func(http.Handler) http.Handler {
	sessionChain := RequireSession(p)
	authzChain := RequireAnyPermission(permission.OrgDatasource)

	return func(next http.Handler) http.Handler {
		protected := sessionChain(authzChain(next))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerKey := r.Header.Get(SyncTriggerAPIKeyHeader)
			if p.Cfg.SyncTriggerAPIKey != "" && headerKey != "" &&
				subtle.ConstantTimeCompare([]byte(headerKey), []byte(p.Cfg.SyncTriggerAPIKey)) == 1 {
				companyIDStr := r.Header.Get(CompanyIDHeader)
				if companyIDStr == "" {
					httputil.WriteStatus(w, http.StatusBadRequest, "company id required")
					return
				}
				companyID, err := uuid.Parse(companyIDStr)
				if err != nil {
					httputil.WriteStatus(w, http.StatusBadRequest, "invalid company id")
					return
				}
				companyCtx, err := companySvc.ResolveCompanyContext(r.Context(), companyID)
				if err != nil {
					httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
					return
				}
				ctx := domaincompany.WithContext(r.Context(), companyCtx)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			protected.ServeHTTP(w, r)
		})
	}
}
