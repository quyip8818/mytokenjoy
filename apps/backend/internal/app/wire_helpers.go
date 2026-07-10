package app

import (
	"context"
	"fmt"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func EnqueueWalletSync(st store.Store) func(context.Context, int64) error {
	return func(ctx context.Context, companyID int64) error {
		return st.AsyncJobs().EnqueueWalletSync(
			domaincompany.WithContext(ctx, domaincompany.Context{CompanyID: companyID}),
			companyID,
		)
	}
}

func EnqueueRebalanceCompany(st store.Store) func(context.Context, int64) error {
	return func(ctx context.Context, companyID int64) error {
		return st.AsyncJobs().EnqueueRebalance(ctx, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
	}
}
