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
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/river/workers"
)

type Client struct {
	inner    *river.Client[pgx.Tx]
	Enqueuer jobs.Enqueuer
}

type Deps struct {
	Billing          domainbilling.Service
	Overrun          domainbudget.OverrunProcessor
	Rebalance        domainbudget.Rebalancer
	NewAPISync       newapisync.OutboxHandler
	OrgSync          domainorg.SyncService
	MonthlyRebalance *domainbudget.MonthlyRebalanceScheduler
}

func NewClient(cfg config.Config, pool *pgxpool.Pool, deps Deps, logger *slog.Logger) (*Client, error) {
	if logger == nil {
		logger = slog.Default()
	}
	inner, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues:       queueConfig(cfg),
		Workers:      registerWorkers(deps),
		PeriodicJobs: BuildPeriodicJobs(cfg),
		Logger:       logger,
	})
	if err != nil {
		return nil, fmt.Errorf("new river client: %w", err)
	}
	return &Client{
		inner:    inner,
		Enqueuer: jobs.NewEnqueuer(inner),
	}, nil
}

func registerWorkers(deps Deps) *river.Workers {
	workersBundle := river.NewWorkers()
	river.AddWorker(workersBundle, workers.NewWalletSyncWorker(deps.Billing))
	river.AddWorker(workersBundle, workers.NewRebalanceWorker(deps.Rebalance))
	river.AddWorker(workersBundle, workers.NewOverrunWorker(deps.Overrun))
	river.AddWorker(workersBundle, workers.NewNewAPISyncWorker(deps.NewAPISync))
	river.AddWorker(workersBundle, workers.NewOrgSyncWorker(deps.OrgSync))
	river.AddWorker(workersBundle, workers.NewMonthlyRebalanceWorker(deps.MonthlyRebalance))
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
