package workers

import (
	"context"

	"github.com/riverqueue/river"
	domainpricing "github.com/tokenjoy/backend/internal/domain/pricing"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

// PricingFullSyncWorker pushes all global prices to NewAPI (ratio cache alignment).
type PricingFullSyncWorker struct {
	river.WorkerDefaults[jobs.PricingFullSyncArgs]
	pricingSvc *domainpricing.Service
}

func NewPricingFullSyncWorker(pricingSvc *domainpricing.Service) *PricingFullSyncWorker {
	return &PricingFullSyncWorker{pricingSvc: pricingSvc}
}

func (w *PricingFullSyncWorker) Work(ctx context.Context, _ *river.Job[jobs.PricingFullSyncArgs]) error {
	return w.pricingSvc.FullSyncToNewAPI(ctx)
}
