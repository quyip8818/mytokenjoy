package workers

import (
	"context"
	"errors"

	"github.com/riverqueue/river"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type WalletSyncWorker struct {
	river.WorkerDefaults[jobs.WalletSyncArgs]
	billing domainbilling.Service
}

func NewWalletSyncWorker(billing domainbilling.Service) *WalletSyncWorker {
	return &WalletSyncWorker{billing: billing}
}

func (w *WalletSyncWorker) Work(ctx context.Context, job *river.Job[jobs.WalletSyncArgs]) error {
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	err := w.billing.SyncCompanyWallet(entryCtx, job.Args.CompanyID)
	if errors.Is(err, domainbilling.ErrWalletNotConfigured) {
		return river.JobCancel(err)
	}
	return cancelIfNonRetryable(err)
}
