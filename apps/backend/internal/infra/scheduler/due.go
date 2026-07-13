package scheduler

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

const reconcileStaleWindow = 7 * 24 * time.Hour

type DueWork struct {
	CompanyID               int64
	NeedsOrgSync            bool
	NeedsMonthRebalance     bool
	NeedsBudgetProject      bool
	NeedsBudgetReconcile    bool
	NeedsDashboardProject   bool
	NeedsDashboardReconcile bool
}

type Service struct {
	cfg   config.Config
	store store.Store
}

func NewService(cfg config.Config, st store.Store) *Service {
	return &Service{cfg: cfg, store: st}
}

func (s *Service) CollectDue(ctx context.Context, now time.Time) ([]DueWork, error) {
	var due []DueWork
	err := company.ForEachActiveCompany(ctx, s.store.Company(), func(entryCtx context.Context, co store.Company) error {
		work, ok, err := s.tenantDue(entryCtx, co.ID, now)
		if err != nil {
			return err
		}
		if ok {
			due = append(due, work)
		}
		return nil
	})
	return due, err
}

func (s *Service) tenantDue(ctx context.Context, companyID int64, now time.Time) (DueWork, bool, error) {
	work := DueWork{CompanyID: companyID}
	tbs, err := s.store.TenantBackgroundState().Get(ctx, companyID)
	if err != nil {
		return work, false, err
	}

	orgSyncNeeded, err := s.orgSyncDue(ctx, tbs, now)
	if err != nil {
		return work, false, err
	}
	if orgSyncNeeded {
		hasPending, err := s.store.RiverJob().HasActiveOrgSync(ctx, companyID)
		if err != nil {
			return work, false, err
		}
		if !hasPending {
			work.NeedsOrgSync = true
		}
	}

	currentMonth := pkgbudget.OpenSnapshotKey(pkgbudget.PeriodMonthly, s.cfg.Clock()).String()
	if monthDue(tbs, currentMonth) {
		work.NeedsMonthRebalance = true
	}

	budgetLag, err := s.projectionLag(ctx, store.BudgetProjectionStream)
	if err != nil {
		return work, false, err
	}
	if budgetLag {
		work.NeedsBudgetProject = true
	}
	if budgetReconcileDue(tbs, budgetLag, now) {
		work.NeedsBudgetReconcile = true
	}

	dashboardLag, err := s.projectionLag(ctx, store.DashboardProjectionStream)
	if err != nil {
		return work, false, err
	}
	if dashboardLag {
		work.NeedsDashboardProject = true
	}
	if dashboardReconcileDue(tbs, dashboardLag, now) {
		work.NeedsDashboardReconcile = true
	}

	ok := work.NeedsOrgSync || work.NeedsMonthRebalance || work.NeedsBudgetProject ||
		work.NeedsBudgetReconcile || work.NeedsDashboardProject || work.NeedsDashboardReconcile
	return work, ok, nil
}

func (s *Service) orgSyncDue(ctx context.Context, tbs *store.TenantBackgroundState, now time.Time) (bool, error) {
	if orgDue(tbs, now) {
		return true, nil
	}
	if !orgScheduleMissing(tbs) {
		return false, nil
	}
	integration, err := s.store.Org().Integration(ctx)
	if err != nil {
		return false, err
	}
	return integration.ToSyncConfig().Enabled, nil
}

func orgDue(tbs *store.TenantBackgroundState, now time.Time) bool {
	if tbs == nil || tbs.NextOrgSyncAt == nil {
		return false
	}
	return !tbs.NextOrgSyncAt.After(now)
}

func orgScheduleMissing(tbs *store.TenantBackgroundState) bool {
	return tbs == nil || tbs.NextOrgSyncAt == nil
}

func monthDue(tbs *store.TenantBackgroundState, currentMonth string) bool {
	if tbs == nil {
		return true
	}
	return tbs.LastRebalancedPeriod != currentMonth
}

func budgetReconcileDue(tbs *store.TenantBackgroundState, lag bool, now time.Time) bool {
	if lag {
		return false
	}
	if tbs == nil || tbs.LastBudgetReconcileAt == nil {
		return true
	}
	return now.Sub(*tbs.LastBudgetReconcileAt) >= reconcileStaleWindow
}

func dashboardReconcileDue(tbs *store.TenantBackgroundState, lag bool, now time.Time) bool {
	if lag {
		return false
	}
	if tbs == nil || tbs.LastDashboardReconcileAt == nil {
		return true
	}
	return now.Sub(*tbs.LastDashboardReconcileAt) >= reconcileStaleWindow
}

func (s *Service) projectionLag(ctx context.Context, stream string) (bool, error) {
	var progress *store.ProjectionProgress
	var err error
	switch stream {
	case store.BudgetProjectionStream:
		progress, err = s.store.BudgetProjectionProgress().Get(ctx, stream)
	case store.DashboardProjectionStream:
		progress, err = s.store.DashboardProjectionProgress().Get(ctx, stream)
	default:
		return false, nil
	}
	if err != nil {
		return false, err
	}
	cursor := store.LedgerProjectorCursor{
		LastOccurredAt: progress.LastOccurredAt,
		LastLedgerID:   progress.LastLedgerID,
		Limit:          1,
	}
	batch, err := s.store.Ledger().ListCallSettledAfterCursor(ctx, cursor)
	if err != nil {
		return false, err
	}
	return len(batch) > 0, nil
}
