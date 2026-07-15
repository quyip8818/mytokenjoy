package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type BulkEnqueuer struct {
	cfg      config.Config
	enqueuer jobs.Enqueuer
}

func NewBulkEnqueuer(cfg config.Config, enqueuer jobs.Enqueuer) *BulkEnqueuer {
	return &BulkEnqueuer{cfg: cfg, enqueuer: enqueuer}
}

func (b *BulkEnqueuer) EnqueueDue(ctx context.Context, st store.Store, work []DueWork, now time.Time) error {
	batchSize := b.cfg.WatchdogBulkBatchSize()
	for start := 0; start < len(work); start += batchSize {
		end := start + batchSize
		if end > len(work) {
			end = len(work)
		}
		for _, item := range work[start:end] {
			entryCtx := company.WithDefaultCompany(ctx, item.CompanyID)
			if err := b.enqueueTenant(entryCtx, st, item, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *BulkEnqueuer) enqueueTenant(ctx context.Context, st store.Store, item DueWork, now time.Time) error {
	if item.NeedsOrgSync {
		tbs, err := st.TenantBackgroundState().Get(ctx, item.CompanyID)
		if err != nil {
			return err
		}
		at := now
		if tbs != nil && tbs.NextOrgSyncAt != nil {
			at = *tbs.NextOrgSyncAt
		}
		if err := jobs.InsertOrgSync(ctx, b.enqueuer, nil, item.CompanyID, &at); err != nil {
			return fmt.Errorf("org_sync company %d: %w", item.CompanyID, err)
		}
	}
	if item.NeedsMonthRebalance {
		axisID := fmt.Sprintf("%d", item.CompanyID)
		if err := jobs.InsertRebalance(ctx, b.enqueuer, nil, item.CompanyID, store.RebalanceAxisCompany, axisID); err != nil {
			return fmt.Errorf("rebalance company %d: %w", item.CompanyID, err)
		}
	}
	if item.NeedsBudgetReconcile {
		if err := jobs.InsertBudgetReconcile(ctx, b.enqueuer, nil, item.CompanyID); err != nil {
			return fmt.Errorf("budget_reconcile company %d: %w", item.CompanyID, err)
		}
	}
	if item.NeedsDashboardProject {
		if err := jobs.InsertDashboardProject(ctx, b.enqueuer, nil, item.CompanyID); err != nil {
			return fmt.Errorf("dashboard_project company %d: %w", item.CompanyID, err)
		}
	}
	if item.NeedsDashboardReconcile {
		if err := jobs.InsertDashboardReconcile(ctx, b.enqueuer, nil, item.CompanyID); err != nil {
			return fmt.Errorf("dashboard_reconcile company %d: %w", item.CompanyID, err)
		}
	}
	return nil
}
