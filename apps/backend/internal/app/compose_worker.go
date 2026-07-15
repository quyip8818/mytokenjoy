// compose_worker.go — River + Ingest background workers.
// Budget/Dashboard Projector + Reconcile are constructed here only (not in HTTP domain services).
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/ingest"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/infra/scheduler"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
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

func logPoolFromStore(st store.Store) *pgxpool.Pool {
	type logPoolStore interface {
		LogPool() *pgxpool.Pool
	}
	if p, ok := st.(logPoolStore); ok {
		return p.LogPool()
	}
	return nil
}

type backgroundWorkers struct {
	ingest *ingest.Worker
	river  *riverinfra.Client
}

func buildBackgroundWorkers(cfg config.Config, logger *slog.Logger, st store.Store, reg ServiceRegistry, holder *jobs.Holder, orgAdmin *OrgRiverAdminHolder) (*backgroundWorkers, error) {
	pool := postgresPool(st)
	if pool == nil {
		return nil, fmt.Errorf("postgres pool unavailable")
	}

	budgetEnqueuer := NewBudgetEnqueuer(holder)
	budgetCache := budgetcheck.WrapStore(reg.Infra.budgetCheck)
	budgetAsync := domainbudget.NewAsync(cfg, st, budgetEnqueuer, budgetCache, logger, domainbudget.WithProjectorNotifier(reg.Infra.notifier))
	dashboardProjector := domaindashboard.NewProjector(cfg, st, NewDashboardEnqueuer(holder), logger)
	dashboardReconcile := domaindashboard.NewReconcileService(cfg, st, NewDashboardEnqueuer(holder), logger)
	sched := scheduler.NewService(cfg, st)
	bulk := scheduler.NewBulkEnqueuer(cfg, holder)

	riverClient, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{
		Cfg:                cfg,
		Store:              st,
		Billing:            reg.BillingSvc,
		Overrun:            reg.Overrun,
		Rebalance:          reg.Rebalance,
		NewAPISync:         reg.Infra.newAPISync,
		OrgSync:            reg.OrgSync,
		BudgetProjector:    budgetAsync.Projector,
		BudgetReconcile:    budgetAsync.Reconcile,
		DashboardProjector: dashboardProjector,
		DashboardReconcile: dashboardReconcile,
		Scheduler:          sched,
		BulkEnqueuer:       bulk,
		DisablePeriodic:    !cfg.RiverPeriodicEnabled,
	}, logger)
	if err != nil {
		return nil, err
	}
	if cfg.RiverEnabled {
		holder.Set(riverClient.Enqueuer)
		if orgAdmin != nil {
			orgAdmin.Set(riverClient)
		}
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

	// Wire LISTEN/NOTIFY for low-latency ingest wake-up.
	if logPool := logPoolFromStore(st); logPool != nil {
		listener, err := postgres.NewPGListener(context.Background(), logPool)
		if err == nil {
			ingestWorker.SetListener(listener)
		} else {
			logger.Warn("failed to create ingest LISTEN connection", "error", err)
		}
	}

	return &backgroundWorkers{
		ingest: ingestWorker,
		river:  riverClient,
	}, nil
}

func (b *backgroundWorkers) start(ctx context.Context, cfg config.Config) {
	if b == nil {
		return
	}
	if b.ingest != nil && cfg.IngestEnabled() {
		b.ingest.Start(ctx)
	}
	if b.river != nil && cfg.RiverEnabled {
		go func() {
			if err := b.river.Start(ctx); err != nil && ctx.Err() == nil {
				slog.Error("river client stopped", "error", err)
			}
		}()
	}
}

func (b *backgroundWorkers) stop(ctx context.Context) {
	if b == nil {
		return
	}
	if b.river != nil {
		_ = b.river.Stop(ctx)
	}
}

var _ = (*pgxpool.Pool)(nil)
