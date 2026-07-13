package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

type CompanyService interface {
	ResolveCompanyContext(ctx context.Context, companyID int64) (company.Context, error)
}

func CompanyResolve(cfg config.Config, companySvc CompanyService, tokenIssuer sessiontoken.Issuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/platform/") {
				next.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/internal/") {
				next.ServeHTTP(w, r)
				return
			}

			companyID := cfg.LocalCompanyID
			if claims, ok := httpx.ResolveMemberClaims(r, tokenIssuer); ok && claims.CompanyID > 0 {
				companyID = claims.CompanyID
			}

			companyCtx, err := companySvc.ResolveCompanyContext(r.Context(), companyID)
			if err != nil {
				if domain.IsNotFound(err) {
					httputil.WriteStatus(w, http.StatusBadRequest, "Company not found")
					return
				}
				httputil.WriteError(w, err)
				return
			}
			if companyCtx.Status == "" {
				companyCtx.Status = "active"
			}
			ctx := company.WithContext(r.Context(), companyCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
