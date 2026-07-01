package testutil

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store/seed"
)

func Ctx() context.Context {
	return company.DefaultContext(seed.DefaultCompanyID)
}
