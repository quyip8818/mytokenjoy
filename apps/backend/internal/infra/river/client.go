package riverinfra

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/infra/river/periodic"
	"github.com/tokenjoy/backend/internal/infra/river/workers"
	"github.com/tokenjoy/backend/internal/infra/scheduler"
	"github.com/tokenjoy/backend/internal/store"
)

type Client struct {
	inner    *river.Client[pgx.Tx]
	Enqueuer jobs.Enqueuer
	store    store.Store
}

type Deps struct {
	Cfg                  config.Config
	Store                store.Store
	Billing              domainbilling.Service
	Overrun              domainbudget.OverrunProcessor
	Rebalance            domainbudget.Rebalancer
	NewAPISync           newapisync.OutboxHandler
	OrgSync              domainorg.SyncService
	BudgetProjector      *domainbudget.Projector
	BudgetReconcile      *domainbudget.ReconcileService
	DashboardProjector   *domaindashboard.Projector
	DashboardReconcile   *domaindashboard.ReconcileService
	Scheduler            *scheduler.Service
	BulkEnqueuer         *scheduler.BulkEnqueuer
	NotificationRegistry *notification.Registry
	DisablePeriodic      bool // tests: skip tenant_watchdog periodic registration
}

func NewClient(cfg config.Config, pool *pgxpool.Pool, deps Deps, logger *slog.Logger) (*Client, error) {
	logger = quietLogger(logger)
	var periodicJobs []*river.PeriodicJob
	if !deps.DisablePeriodic {
		periodicJobs = periodic.BuildWatchdogJobs(cfg)
	}
	inner, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues:       queueConfig(cfg),
		Workers:      registerWorkers(deps),
		PeriodicJobs: periodicJobs,
		Logger:       logger,
	})
	if err != nil {
		return nil, fmt.Errorf("new river client: %w", err)
	}
	return &Client{
		inner:    inner,
		Enqueuer: jobs.NewEnqueuer(inner),
		store:    deps.Store,
	}, nil
}

func registerWorkers(deps Deps) *river.Workers {
	workersBundle := river.NewWorkers()
	river.AddWorker(workersBundle, workers.NewWalletSyncWorker(deps.Billing))
	river.AddWorker(workersBundle, workers.NewRebalanceWorker(deps.Rebalance, deps.Store, deps.Cfg))
	river.AddWorker(workersBundle, workers.NewOverrunWorker(deps.Overrun))
	river.AddWorker(workersBundle, workers.NewNewAPISyncWorker(deps.NewAPISync))
	river.AddWorker(workersBundle, workers.NewOrgSyncWorker(deps.OrgSync))
	river.AddWorker(workersBundle, workers.NewBudgetProjectionWorker(deps.BudgetProjector))
	river.AddWorker(workersBundle, workers.NewBudgetReconcileWorker(deps.BudgetReconcile, deps.Store))
	river.AddWorker(workersBundle, workers.NewDashboardProjectWorker(deps.DashboardProjector))
	river.AddWorker(workersBundle, workers.NewDashboardReconcileWorker(deps.DashboardReconcile, deps.Store))
	river.AddWorker(workersBundle, workers.NewWatchdogWorker(deps.Scheduler, deps.BulkEnqueuer, deps.Store, deps.Cfg))
	if deps.NotificationRegistry != nil {
		river.AddWorker(workersBundle, workers.NewNotificationDeliveryWorker(deps.NotificationRegistry))
	}
	return workersBundle
}

func queueConfig(cfg config.Config) map[string]river.QueueConfig {
	maxWorkers := cfg.RiverMaxWorkers()
	return map[string]river.QueueConfig{
		config.RiverQueueCritical: {MaxWorkers: max(1, maxWorkers*2/5)},
		config.RiverQueueDefault:  {MaxWorkers: max(1, maxWorkers*2/5)},
		config.RiverQueueLow:      {MaxWorkers: max(1, maxWorkers/5)},
	}
}

func (c *Client) Start(ctx context.Context) error {
	if c == nil || c.inner == nil {
		return nil
	}
	return c.inner.Start(ctx)
}

func (c *Client) Stop(ctx context.Context) error {
	if c == nil || c.inner == nil {
		return nil
	}
	return c.inner.Stop(ctx)
}

func (c *Client) Inner() *river.Client[pgx.Tx] {
	if c == nil {
		return nil
	}
	return c.inner
}

func (c *Client) CancelOrgSyncPending(ctx context.Context, companyID int64) error {
	if c == nil || c.inner == nil || c.store == nil {
		return nil
	}
	ids, err := c.store.RiverJob().ListCancellableOrgSyncJobIDs(ctx, companyID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := c.inner.JobCancel(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
