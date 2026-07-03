package middleware

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/store"
)

type authzRevisionStore interface {
	Company() store.CompanyRepository
}

func AuthzRevisionHeader(st authzRevisionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var revision int64
			var hasRevision bool
			if sessionCtx, ok := SessionFromContext(r.Context()); ok {
				revision = sessionCtx.AuthzRevision
				hasRevision = true
			} else if companyID := ctxcompany.ID(r.Context()); companyID > 0 {
				if rev, err := st.Company().GetAuthzRevision(r.Context(), companyID); err == nil {
					revision = rev
					hasRevision = true
				}
			}
			if hasRevision {
				httpx.SetAuthzRevisionHeader(w, revision)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func WithAuthzRevisionOnWriter(w http.ResponseWriter, revision int64) {
	httpx.SetAuthzRevisionHeader(w, revision)
}
