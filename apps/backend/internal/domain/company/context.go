package company

import (
	"context"

	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
)

type Context = ctxcompany.Info

func WithContext(ctx context.Context, info Context) context.Context {
	return ctxcompany.With(ctx, info)
}

func FromContext(ctx context.Context) (Context, bool) {
	return ctxcompany.From(ctx)
}

func CompanyID(ctx context.Context) int64 {
	return ctxcompany.ID(ctx)
}
