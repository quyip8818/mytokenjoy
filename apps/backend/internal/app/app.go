package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/infra/worker"
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
	skipWorker bool
}

type Option func(*options)

func WithoutWorker() Option {
	return func(o *options) {
		o.skipWorker = true
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

	infraDeps, err := buildInfraWithStore(cfg, logger, st)
	if err != nil {
		return nil, err
	}

	registry := buildServiceRegistry(cfg, infraDeps, buildDomainServices(cfg, infraDeps, logger))
	if err := registry.Platform.BootstrapIfNeeded(ctx); err != nil {
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

func (a *App) Close() {
	for _, closer := range a.closers {
		closer()
	}
}
