package middleware

import (
	"context"
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/infra/platformauth"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

type PlatformService = platformauth.Service

type platformContextKey struct{}

func WithPlatformOperator(ctx context.Context, operatorID string) context.Context {
	return context.WithValue(ctx, platformContextKey{}, operatorID)
}

func PlatformOperatorFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(platformContextKey{}).(string)
	return id, ok
}

func PlatformAuth(cfg config.Config, svc PlatformService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/platform/auth/login" && r.Method == http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			operatorID := common.ResolvePlatformOperatorID(r)
			if operatorID == "" {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}
			ctx := WithPlatformOperator(r.Context(), operatorID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
