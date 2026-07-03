package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
)

func SessionFromContext(ctx context.Context) (types.SessionContext, bool) {
	return httpx.SessionFromContext(ctx)
}

func RequireSession(p httpdeps.Protected) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := httpx.ParseMemberToken(r, p.SessionToken)
			if err != nil {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}

			sessionCtx, err := p.AuthzSvc.GetSessionContext(r.Context(), claims.CompanyID, claims.Subject)
			if err != nil {
				var domainErr *domain.DomainError
				if errors.As(err, &domainErr) && domainErr.Status == domain.StatusNotFound {
					httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
					return
				}
				httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
				return
			}
			if sessionCtx.CompanyID != claims.CompanyID || sessionCtx.Member.ID != claims.Subject {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}
			if sessionCtx.Member.Status != types.MemberStatusActive {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}

			ctx := httpx.WithSessionClaims(r.Context(), claims)
			ctx = httpx.WithSessionContext(ctx, sessionCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
