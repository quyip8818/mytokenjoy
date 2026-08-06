package httpx

import (
	"context"
	"net/http"
	"strconv"

	"github.com/tokenjoy/backend/internal/domain/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type sessionClaimsKey struct{}

func WithSessionClaims(ctx context.Context, claims sessiontoken.Claims) context.Context {
	return context.WithValue(ctx, sessionClaimsKey{}, claims)
}

func SessionClaimsFromContext(ctx context.Context) (sessiontoken.Claims, bool) {
	claims, ok := ctx.Value(sessionClaimsKey{}).(sessiontoken.Claims)
	return claims, ok
}

// WithSessionContext delegates to types.WithSessionContext.
func WithSessionContext(ctx context.Context, sessionCtx types.SessionContext) context.Context {
	return types.WithSessionContext(ctx, sessionCtx)
}

// SessionFromContext delegates to types.SessionFromContext.
func SessionFromContext(ctx context.Context) (types.SessionContext, bool) {
	return types.SessionFromContext(ctx)
}

func SetAuthzRevisionHeader(w http.ResponseWriter, revision int64) {
	w.Header().Set(HeaderAuthzRevision, strconv.FormatInt(revision, 10))
}
