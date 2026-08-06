package company

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/support/tenant"
)

type Context = tenant.Info

func WithContext(ctx context.Context, info Context) context.Context {
	return tenant.With(ctx, info)
}

func FromContext(ctx context.Context) (Context, bool) {
	return tenant.From(ctx)
}

func CompanyID(ctx context.Context) uuid.UUID {
	return tenant.ID(ctx)
}

func DefaultContext(companyID uuid.UUID) context.Context {
	return WithContext(context.Background(), Context{CompanyID: companyID, Status: "active"})
}

func WithDefaultCompany(ctx context.Context, companyID uuid.UUID) context.Context {
	if _, ok := FromContext(ctx); ok {
		return ctx
	}
	return WithContext(ctx, Context{CompanyID: companyID, Status: "active"})
}
