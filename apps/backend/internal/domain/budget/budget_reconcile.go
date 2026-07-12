package budget

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

type ReconcileService struct {
	cfg          config.Config
	store        store.Store
	enqueuer     jobs.Enqueuer
	logger       *slog.Logger
	gatewayCache budgetcheck.Store
}

func (s *ReconcileService) RunCompany(ctx context.Context, companyID int64) error {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if co == nil {
		return nil
	}
	ctx = company.WithContext(ctx, companyFromStore(*co))

	since := reconcileWindowStart(clock.NowUTC(s.cfg.Clock()))
	entries, err := s.store.Ledger().ListCallSettledSince(ctx, since)
	if err != nil {
		return err
	}
	nodes := s.store.Org().Nodes()
	expected, err := ExpectedConsumed(ctx, nodes, entries, s.cfg.Clock())
	if err != nil {
		return err
	}

	var summaries []store.GatewaySoftSummary
	repaired := false
	repairedAxes := make(map[AxisKey]struct{})
	err = s.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		consumedRepo := tx.BudgetConsumed()
		for key, want := range expected {
			got, found, err := consumedRepo.GetConsumed(ctx, key.Kind, key.AxisID, key.PeriodKey)
			if err != nil {
				return err
			}
			if !found {
				got = 0
			}
			if !ConsumedDrift(want, got) {
				continue
			}
			if err := consumedRepo.SetConsumed(ctx, key.Kind, key.AxisID, key.PeriodKey, want); err != nil {
				return err
			}
			repaired = true
			repairedAxes[key] = struct{}{}
			if s.logger != nil {
				s.logger.Warn("budget reconcile drift repaired",
					"company_id", companyID,
					"axis_kind", key.Kind,
					"axis_id", key.AxisID,
					"period_key", key.PeriodKey,
					"expected", want,
					"actual", got,
				)
			}
		}
		if !repaired {
			return nil
		}
		affectedKeys, err := AffectedPlatformKeyIDs(ctx, tx, repairedAxes)
		if err != nil {
			return err
		}
		updates, err := ComputeGatewaySummaryUpdates(ctx, tx, affectedKeys, s.cfg.Clock())
		if err != nil {
			return err
		}
		if len(updates) > 0 {
			summaries, err = tx.GatewaySoftSummaries().UpdateBatch(ctx, updates)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !repaired {
		return nil
	}

	budgetcheck.RefreshSummaries(ctx, s.gatewayCache, s.logger, companyID, summaries)
	return jobs.InsertRebalance(ctx, s.enqueuer, nil, companyID, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
}

func (s *ReconcileService) FanoutReconcileJobs(ctx context.Context) error {
	return company.ForEachActiveCompany(ctx, s.store.Company(), func(entryCtx context.Context, co store.Company) error {
		return jobs.InsertBudgetReconcile(entryCtx, s.enqueuer, nil, co.ID)
	})
}

func reconcileWindowStart(now time.Time) time.Time {
	return now.AddDate(0, -2, 0)
}
