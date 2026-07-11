package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/infra/ingest"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/store"
)

func postgresPool(st store.Store) *pgxpool.Pool {
	type poolStore interface {
		Pool() *pgxpool.Pool
	}
	if p, ok := st.(poolStore); ok {
		return p.Pool()
	}
	return nil
}

type backgroundWorkers struct {
	ingest *ingest.Worker
	river  *riverinfra.Client
}

func buildBackgroundWorkers(cfg config.Config, logger *slog.Logger, st store.Store, reg ServiceRegistry, holder *jobs.Holder) (*backgroundWorkers, error) {
	pool := postgresPool(st)
	if pool == nil {
		return nil, fmt.Errorf("postgres pool unavailable")
	}

	budgetAsync := domainbudget.NewAsync(cfg, st, holder, reg.Infra.budgetCheck, logger)
	riverClient, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{
		Billing:            reg.BillingSvc,
		Overrun:            reg.Overrun,
		Rebalance:          reg.Rebalance,
		NewAPISync:         reg.Infra.newAPISync,
		OrgSync:            reg.OrgSync,
		MonthlyRebalance:   domainbudget.NewMonthlyRebalanceScheduler(cfg, st, holder),
		BudgetProjector:    budgetAsync.Projector,
		BudgetReconcile:    budgetAsync.Reconcile,
		DashboardProjector: domaindashboard.NewProjector(cfg, st, holder, logger),
		DashboardReconcile: domaindashboard.NewReconcileService(cfg, st, holder, logger),
	}, logger)
	if err != nil {
		return nil, err
	}
	if cfg.RiverEnabled {
		holder.Set(riverClient.Enqueuer)
	}

	ingestWorker := ingest.NewWorker(
		cfg,
		st.Logs(),
		reg.IngestSvc,
		reg.IngestQueue,
		reg.IngestMetrics,
		st.SchedulerLock(),
		reg.BillingSvc,
		logger,
	)

	return &backgroundWorkers{
		ingest: ingestWorker,
		river:  riverClient,
	}, nil
}

func (b *backgroundWorkers) start(ctx context.Context, cfg config.Config) {
	if b == nil {
		return
	}
	if cfg.IngestEnabled() && b.ingest != nil {
		b.ingest.Start(ctx)
	}
	if cfg.RiverEnabled && b.river != nil {
		if err := b.river.Start(ctx); err != nil {
			slog.Default().Error("river client start failed", "error", err)
		}
	}
}

func (b *backgroundWorkers) stop(ctx context.Context) {
	if b == nil || b.river == nil {
		return
	}
	_ = b.river.Stop(ctx)
}
