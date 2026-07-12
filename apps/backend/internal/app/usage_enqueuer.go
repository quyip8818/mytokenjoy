package app

import (
	"context"

	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type usageIngestEnqueuer struct {
	enqueuer jobs.Enqueuer
}

func NewUsageIngestEnqueuer(enqueuer jobs.Enqueuer) domainusage.IngestJobEnqueuer {
	return usageIngestEnqueuer{enqueuer: jobsOrNoop(enqueuer)}
}

func (u usageIngestEnqueuer) EnqueueAfterIngest(ctx context.Context, tx store.Tx, companyID int64) error {
	if err := jobs.InsertBudgetProject(ctx, u.enqueuer, tx, companyID); err != nil {
		return err
	}
	return jobs.InsertWalletSync(ctx, u.enqueuer, tx, companyID)
}

var _ domainusage.IngestJobEnqueuer = usageIngestEnqueuer{}
