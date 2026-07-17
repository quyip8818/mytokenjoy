package ctxcompany

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

type Info struct {
	CompanyID          uuid.UUID
	NewAPIWalletUserID int64
	Type               string
	Status             string
}

func With(ctx context.Context, info Info) context.Context {
	return context.WithValue(ctx, contextKey{}, info)
}

func From(ctx context.Context) (Info, bool) {
	info, ok := ctx.Value(contextKey{}).(Info)
	return info, ok
}

func ID(ctx context.Context) uuid.UUID {
	if info, ok := From(ctx); ok {
		return info.CompanyID
	}
	return uuid.Nil
}
