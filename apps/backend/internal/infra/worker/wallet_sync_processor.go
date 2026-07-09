package worker

import (
	"context"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/company"
)

func (r *Runner) processWalletSync(ctx context.Context) error {
	entries, err := r.relayJobs.ClaimPendingWalletSync(ctx, 20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entryCtx := r.workerCtx(ctx, entry.CompanyID)
		if err := r.walletSync.SyncCompanyWallet(entryCtx, entry.CompanyID); err != nil {
			r.logger.Warn("wallet_sync failed", "company_id", entry.CompanyID, "error", err)
			continue
		}
		r.markWalletSyncDone(entryCtx, entry.ID)
	}
	return nil
}

func (r *Runner) markWalletSyncDone(ctx context.Context, id string) {
	if err := r.relayJobs.MarkWalletSyncDone(ctx, id); err != nil {
		r.logger.Warn("mark wallet_sync done failed", "id", id, "error", err)
	}
}

type billingWalletSync struct {
	svc domainbilling.Service
}

func (b billingWalletSync) ReconcileWalletDrift(ctx context.Context) error {
	return b.svc.ReconcileWalletDrift(ctx)
}

func (b billingWalletSync) SyncCompanyWallet(ctx context.Context, companyID int64) error {
	return b.svc.SyncCompanyWallet(company.WithContext(ctx, company.Context{CompanyID: companyID}), companyID)
}

func (r *Runner) processWalletReconcile(ctx context.Context) error {
	return r.walletSync.ReconcileWalletDrift(ctx)
}
