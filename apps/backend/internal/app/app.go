package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/internal/worker"
)

func openStore(ctx context.Context, cfg config.Config) (store.Store, error) {
	snapshot := seed.Load(cfg)
	var st store.Store
	var err error
	if cfg.DatabaseURL == "" {
		st = store.NewMemory(snapshot)
	} else {
		st, err = postgres.New(ctx, cfg, snapshot)
		if err != nil {
			return nil, err
		}
	}
	if cfg.IsDemoProfile() {
		if err := seed.ApplyUsageBuckets(ctx, st, cfg); err != nil {
			if closer, ok := st.(interface{ Close() }); ok {
				closer.Close()
			}
			return nil, err
		}
	}
	return st, nil
}

type App struct {
	Config  config.Config
	Store   store.Store
	Router  http.Handler
	Worker  *worker.Runner
	closers []func()
}

type options struct {
	skipWorker bool
}

type Option func(*options)

func WithoutWorker() Option {
	return func(o *options) {
		o.skipWorker = true
	}
}

func New(cfg config.Config, logger *slog.Logger, opts ...Option) (*App, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	ctx := context.Background()
	infraDeps, err := buildInfra(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}
	st := infraDeps.store

	services := buildDomainServices(cfg, infraDeps, logger)
	runner := worker.NewRunner(cfg, st, infraDeps.adminClient, infraDeps.lifecycle, services.ingest, services.rebalance, services.org, logger)

	router := httpapi.NewRouter(httpapi.Deps{
		Config:       cfg,
		Logger:       logger,
		SessionSvc:   services.session,
		OrgSvc:       services.org,
		BudgetSvc:    services.budget,
		KeysSvc:      services.keys,
		ModelsSvc:    services.models,
		DashboardSvc: services.dashboard,
		AuditSvc:     services.audit,
		IngestSvc:    services.ingest,
	})

	workerCtx, cancel := context.WithCancel(context.Background())
	if !o.skipWorker {
		runner.Start(workerCtx)
	}

	return &App{
		Config: cfg,
		Store:  st,
		Router: router,
		Worker: runner,
		closers: []func(){
			cancel,
			func() {
				if closer, ok := st.(interface{ Close() }); ok {
					closer.Close()
				}
			},
		},
	}, nil
}

func (a *App) Close() {
	for _, closer := range a.closers {
		closer()
	}
}
