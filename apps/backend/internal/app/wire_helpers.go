package app

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
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

func EnqueueRebalanceCompany(cfg config.Config, st store.Store) func(context.Context, int64) error {
	return func(ctx context.Context, companyID int64) error {
		if !cfg.NewAPIEnabled {
			return nil
		}
		return st.AsyncJobs().EnqueueRebalance(
			domaincompany.WithContext(ctx, domaincompany.Context{CompanyID: companyID}),
			store.RebalanceAxisCompany,
			fmt.Sprintf("%d", companyID),
		)
	}
}

func EnqueueRebalanceAxis(cfg config.Config, st store.Store) func(context.Context, string, string) error {
	return func(ctx context.Context, axisKind, axisID string) error {
		if !cfg.NewAPIEnabled {
			return nil
		}
		return st.AsyncJobs().EnqueueRebalance(ctx, axisKind, axisID)
	}
}
