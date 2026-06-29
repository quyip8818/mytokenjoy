package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

type sessionContextKey struct{}

func WithSessionContext(ctx context.Context, sessionCtx types.SessionContext) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, sessionCtx)
}

func SessionFromContext(ctx context.Context) (types.SessionContext, bool) {
	sessionCtx, ok := ctx.Value(sessionContextKey{}).(types.SessionContext)
	return sessionCtx, ok
}

func RequireSession(cfg config.Config, sessionSvc session.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.IsProdProfile() && common.UsedBearerAuth(r) {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}
			memberID := common.ResolveMemberID(r)
			if memberID == "" {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}

			sessionCtx, err := sessionSvc.GetByMemberID(memberID)
			if err != nil {
				var domainErr *domain.DomainError
				if errors.As(err, &domainErr) && domainErr.Status == domain.StatusNotFound {
					httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
					return
				}
				httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
				return
			}

			ctx := WithSessionContext(r.Context(), sessionCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
