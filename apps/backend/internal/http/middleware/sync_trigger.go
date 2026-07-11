package middleware

import (
	"crypto/subtle"
	"net/http"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

const SyncTriggerAPIKeyHeader = "X-Sync-API-Key"
const CompanySlugHeader = "X-Company-Slug"

func AllowSyncTrigger(p httpdeps.Protected, companySvc domaincompany.Service) func(http.Handler) http.Handler {
	sessionChain := RequireSession(p)
	authzChain := RequireAnyPermission(permission.OrgDatasource)

	return func(next http.Handler) http.Handler {
		protected := sessionChain(authzChain(next))
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headerKey := r.Header.Get(SyncTriggerAPIKeyHeader)
			if p.Cfg.SyncTriggerAPIKey != "" && headerKey != "" &&
				subtle.ConstantTimeCompare([]byte(headerKey), []byte(p.Cfg.SyncTriggerAPIKey)) == 1 {
				slug := r.Header.Get(CompanySlugHeader)
				if slug == "" {
					httputil.WriteStatus(w, http.StatusBadRequest, "company slug required")
					return
				}
				companyCtx, err := companySvc.ResolveCompanyContextBySlug(r.Context(), slug)
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
