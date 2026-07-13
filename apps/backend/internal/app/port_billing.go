package app

import (
	"context"
	"fmt"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type billingJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

// NewBillingEnqueuer adapts infra/jobs.Enqueuer to domain/billing.JobEnqueuer.
func NewBillingEnqueuer(enqueuer jobs.Enqueuer) domainbilling.JobEnqueuer {
	return billingJobEnqueuer{enqueuer: jobsOrNoop(enqueuer)}
}

func (b billingJobEnqueuer) InsertWalletSync(ctx context.Context, companyID int64) error {
	return jobs.InsertWalletSync(ctx, b.enqueuer, nil, companyID)
}

func (b billingJobEnqueuer) InsertRebalanceCompany(ctx context.Context, companyID int64) error {
	return jobs.InsertRebalance(ctx, b.enqueuer, nil, companyID, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
}

var _ domainbilling.JobEnqueuer = billingJobEnqueuer{}
