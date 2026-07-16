package adapter

import (
	"context"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type billingJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

// NewBillingEnqueuer adapts infra/jobs.Enqueuer to domain/billing.JobEnqueuer.
func NewBillingEnqueuer(enqueuer jobs.Enqueuer) domainbilling.JobEnqueuer {
	return billingJobEnqueuer{enqueuer: JobsOrNoop(enqueuer)}
}

func (b billingJobEnqueuer) InsertWalletSync(ctx context.Context, companyID int64) error {
	return jobs.InsertWalletSync(ctx, b.enqueuer, nil, companyID)
}

var _ domainbilling.JobEnqueuer = billingJobEnqueuer{}
