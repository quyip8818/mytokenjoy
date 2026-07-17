package httpx

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

type sessionClaimsKey struct{}

func WithSessionClaims(ctx context.Context, claims sessiontoken.Claims) context.Context {
	return context.WithValue(ctx, sessionClaimsKey{}, claims)
}

func SessionClaimsFromContext(ctx context.Context) (sessiontoken.Claims, bool) {
	claims, ok := ctx.Value(sessionClaimsKey{}).(sessiontoken.Claims)
	return claims, ok
}

type sessionContextKey struct{}

func WithSessionContext(ctx context.Context, sessionCtx types.SessionContext) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, sessionCtx)
}

func SessionFromContext(ctx context.Context) (types.SessionContext, bool) {
	sessionCtx, ok := ctx.Value(sessionContextKey{}).(types.SessionContext)
	return sessionCtx, ok
}

type platformContextKey struct{}

func WithPlatformOperator(ctx context.Context, operatorID uuid.UUID) context.Context {
	return context.WithValue(ctx, platformContextKey{}, operatorID)
}

func PlatformOperatorFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(platformContextKey{}).(string)
	return id, ok
}

func SetAuthzRevisionHeader(w http.ResponseWriter, revision int64) {
	w.Header().Set(HeaderAuthzRevision, strconv.FormatInt(revision, 10))
}
