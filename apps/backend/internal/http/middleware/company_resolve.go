package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type CompanyService interface {
	ResolveCompanyContext(ctx context.Context, companyID int64) (company.Context, error)
	ResolveFromMember(ctx context.Context, memberID string) (company.Context, error)
}

func CompanyResolve(cfg config.Config, companySvc CompanyService) func(http.Handler) http.Handler {
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
			var companyID int64
			var companyCtx company.Context
			var err error
			if !cfg.SupportSaas {
				companyID = cfg.DefaultCompanyID
				companyCtx, err = companySvc.ResolveCompanyContext(r.Context(), companyID)
			} else if memberID := common.ResolveMemberID(r); memberID != "" {
				companyCtx, err = companySvc.ResolveFromMember(r.Context(), memberID)
			} else {
				companyID = cfg.DefaultCompanyID
				companyCtx, err = companySvc.ResolveCompanyContext(r.Context(), companyID)
			}
			if err != nil {
				httputil.WriteStatus(w, http.StatusBadRequest, "Company not found")
				return
			}
			if companyCtx.Status == "" {
				companyCtx.Status = store.CompanyStatusActive
			}
			ctx := company.WithContext(r.Context(), companyCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
