package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/infra/worker"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

func openStore(ctx context.Context, cfg config.Config) (store.Store, error) {
	return postgres.New(ctx, cfg)
}

type App struct {
	Config  config.Config
	Store   store.Store
	Router  http.Handler
	Worker  *worker.Runner
	closers []func()
}

type options struct {
	skipWorker  bool
	adminClient newapi.AdminClient
	orgSync     domainorg.SyncService
}

type Option func(*options)

func WithoutWorker() Option {
	return func(o *options) {
		o.skipWorker = true
	}
}

func WithAdminClient(client newapi.AdminClient) Option {
	return func(o *options) {
		o.adminClient = client
	}
}

func WithOrgSync(svc domainorg.SyncService) Option {
	return func(o *options) {
		o.orgSync = svc
	}
}

func New(cfg config.Config, logger *slog.Logger, opts ...Option) (*App, error) {
	ctx := context.Background()
	st, err := openStore(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return newApp(cfg, logger, st, opts...)
}

func newApp(cfg config.Config, logger *slog.Logger, st store.Store, opts ...Option) (*App, error) {
	ctx := context.Background()
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	registry, err := assembleRegistry(cfg, logger, st, o)
	if err != nil {
		return nil, err
	}
	if err := registry.Credentials.BootstrapPlatformIfNeeded(ctx); err != nil {
		return nil, err
	}
	runner := registry.WorkerRunner(logger)

	router := httpapi.NewRouter(registry.HTTPDeps(logger))

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

func assembleRegistry(cfg config.Config, logger *slog.Logger, st store.Store, o options) (ServiceRegistry, error) {
	infraDeps, err := buildInfraWithStore(cfg, logger, st, o.adminClient)
	if err != nil {
		return ServiceRegistry{}, err
	}
	registry := buildServiceRegistry(cfg, infraDeps, buildDomainServices(cfg, infraDeps, logger))
	if o.orgSync != nil {
		registry.OrgSync = o.orgSync
	}
	return registry, nil
}

func (a *App) Close() {
	for _, closer := range a.closers {
		closer()
	}
}
