package app

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

func EnqueueWalletSync(e jobs.Enqueuer) func(context.Context, int64) error {
	return func(ctx context.Context, companyID int64) error {
		return jobs.InsertWalletSync(ctx, e, nil, companyID)
	}
}

func EnqueueRebalanceCompany(e jobs.Enqueuer) func(context.Context, int64) error {
	return func(ctx context.Context, companyID int64) error {
		return jobs.InsertRebalance(ctx, e, nil, companyID, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
	}
}

func EnqueueRebalanceAxis(e jobs.Enqueuer) func(context.Context, string, string) error {
	return func(ctx context.Context, axisKind, axisID string) error {
		return jobs.InsertRebalance(ctx, e, nil, store.CompanyID(ctx), axisKind, axisID)
	}
}
