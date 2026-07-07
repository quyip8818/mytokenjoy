package testutil

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/seed/contract"
)

func Ctx() context.Context {
	return company.DefaultContext(contract.DefaultCompanyID)
}

func CtxForCompany(companyID int64) context.Context {
	return company.DefaultContext(companyID)
}
