package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

func PlatformOperatorFromContext(ctx context.Context) (string, bool) {
	return httpx.PlatformOperatorFromContext(ctx)
}

func PlatformAuth(cfg config.Config, tokenIssuer sessiontoken.Issuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/platform/auth/login" && r.Method == http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := httpx.ParsePlatformToken(r, tokenIssuer)
			if err != nil {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}
			operatorID, parseErr := uuid.Parse(claims.Subject)
			if parseErr != nil {
				httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
				return
			}
			ctx := httpx.WithPlatformOperator(r.Context(), operatorID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
