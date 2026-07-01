package company

import "context"

func DefaultContext(companyID int64) context.Context {
	return WithContext(context.Background(), Context{CompanyID: companyID, Status: "active"})
}

func WithDefaultCompany(ctx context.Context, companyID int64) context.Context {
	if _, ok := FromContext(ctx); ok {
		return ctx
	}
	return WithContext(ctx, Context{CompanyID: companyID, Status: "active"})
}
