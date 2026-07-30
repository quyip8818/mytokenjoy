// compose_worker.go — River background workers.
// Budget Reconcile + Dashboard Projector/Reconcile are constructed here only (not in HTTP domain services).
package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/adapter/enqueue"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	riverinfra "github.com/tokenjoy/backend/internal/infra/river"
	"github.com/tokenjoy/backend/internal/infra/scheduler"
	catalogintegration "github.com/tokenjoy/backend/internal/integration/catalogsync"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/worker/catalogsync"
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
	river               *riverinfra.Client
	catalogSyncExecutor *catalogsync.Executor
	logger              *slog.Logger
}

func buildBackgroundWorkers(cfg config.Config, logger *slog.Logger, st store.Store, reg ServiceRegistry, holder *jobs.Holder, orgAdmin *enqueue.OrgRiverAdminHolder) (*backgroundWorkers, error) {
	pool := postgresPool(st)
	if pool == nil {
		return nil, fmt.Errorf("postgres pool unavailable")
	}

	budgetEnqueuer := enqueue.NewBudgetEnqueuer(holder)
	budgetCache := budgetcheck.WrapStore(reg.Infra.budgetCheck)
	budgetReconcile := domainbudget.NewReconcileService(cfg, st, budgetEnqueuer, budgetCache, logger)
	dashboardProjector := domaindashboard.NewProjector(cfg, st, enqueue.NewDashboardEnqueuer(holder), logger)
	dashboardReconcile := domaindashboard.NewReconcileService(cfg, st, enqueue.NewDashboardEnqueuer(holder), logger)
	sched := scheduler.NewService(cfg, st)
	bulk := scheduler.NewBulkEnqueuer(cfg, holder)

	catalogExecutor := buildCatalogSyncExecutor(cfg, st, reg)

	riverClient, err := riverinfra.NewClient(cfg, pool, riverinfra.Deps{
		Cfg:                  cfg,
		Store:                st,
		LogStore:             st.Logs(),
		Ingest:               reg.IngestSvc,
		Enqueuer:             holder,
		Logger:               logger,
		Overrun:              reg.Overrun,
		Rebalance:            reg.Rebalance,
		NewAPISync:           reg.Infra.newAPISync,
		OrgSync:              reg.OrgSync,
		BudgetReconcile:      budgetReconcile,
		DashboardProjector:   dashboardProjector,
		DashboardReconcile:   dashboardReconcile,
		Scheduler:            sched,
		BulkEnqueuer:         bulk,
		NotificationRegistry: reg.Infra.notificationSvc.Registry(),
		CatalogSyncExecutor:  catalogExecutor,
		DisablePeriodic:      !cfg.RiverPeriodicEnabled,
	}, logger)
	if err != nil {
		return nil, err
	}
	if cfg.RiverEnabled {
		holder.Set(riverClient.Enqueuer)
		if orgAdmin != nil {
			orgAdmin.Set(riverClient)
		}
		if reg.Infra.notificationSvc != nil {
			reg.Infra.notificationSvc.SetEnqueuer(riverClient.Enqueuer)
		}
	}

	return &backgroundWorkers{
		river:               riverClient,
		catalogSyncExecutor: catalogExecutor,
		logger:              logger,
	}, nil
}

func (b *backgroundWorkers) start(ctx context.Context, cfg config.Config) {
	if b == nil {
		return
	}
	if b.river != nil && cfg.RiverEnabled {
		go func() {
			if err := b.river.Start(ctx); err != nil && ctx.Err() == nil {
				slog.Error("river client stopped", "error", err)
			}
		}()
	}
	// ponytail: immediate catalog sync on boot so models are available right after setup.
	// Periodic job handles subsequent syncs. Fire-and-forget in a goroutine to not block startup.
	if b.catalogSyncExecutor != nil {
		go func() {
			if err := b.catalogSyncExecutor.Execute(ctx); err != nil {
				slog.Warn("catalog sync on boot failed", "error", err)
			} else {
				slog.Info("catalog sync on boot completed")
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

// buildCatalogSyncExecutor constructs the catalog sync executor if enabled, nil otherwise.
func buildCatalogSyncExecutor(cfg config.Config, st store.Store, reg ServiceRegistry) *catalogsync.Executor {
	if !cfg.CatalogSyncEnabled {
		return nil
	}
	if cfg.CatalogSyncURL == "" {
		slog.Warn("catalog sync enabled but CATALOG_SYNC_URL empty, skipping")
		return nil
	}
	// Sync token is persisted by setup flow into system_settings.
	syncToken, _ := st.SystemSettings().Get(context.Background(), "catalog_sync_token")
	client := catalogintegration.NewClient(catalogintegration.Config{
		BaseURL:   cfg.CatalogSyncURL,
		SyncToken: syncToken,
	})
	// Local company ID (registered on SaaS) for contract pricing mapping.
	localCoStr, _ := st.SystemSettings().Get(context.Background(), "setup_company_id")
	localCompanyID, _ := uuid.Parse(localCoStr) // zero UUID if not yet set up
	return catalogsync.NewExecutor(client, reg.Infra.adminPort, st, cfg.TokenJoyCompanyID, localCompanyID)
}
