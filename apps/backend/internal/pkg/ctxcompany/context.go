package ctxcompany

import "context"

type contextKey struct{}

type Info struct {
	CompanyID          int64
	Slug               string
	NewAPIWalletUserID int64
	Status             string
}

func With(ctx context.Context, info Info) context.Context {
	return context.WithValue(ctx, contextKey{}, info)
}

func From(ctx context.Context) (Info, bool) {
	info, ok := ctx.Value(contextKey{}).(Info)
	return info, ok
}

func ID(ctx context.Context) int64 {
	if info, ok := From(ctx); ok {
		return info.CompanyID
	}
	return 0
}
