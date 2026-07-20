package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/infra/jobs"
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
	Workers *backgroundWorkers
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

	holder := jobs.NewHolder(jobs.NoopEnqueuer{})
	orgAdmin := adapter.NewOrgRiverAdminHolder(nil)
	registry, err := assembleRegistry(cfg, logger, st, o, holder, orgAdmin)
	if err != nil {
		return nil, err
	}
	if err := registry.Credentials.BootstrapPlatformIfNeeded(ctx); err != nil {
		return nil, err
	}

	bgWorkers, err := buildBackgroundWorkers(cfg, logger, st, registry, holder, orgAdmin)
	if err != nil {
		return nil, err
	}
	router := httpapi.NewRouter(registry.HTTPDeps(logger))

	workerCtx, cancel := context.WithCancel(context.Background())
	if !o.skipWorker {
		bgWorkers.start(workerCtx, cfg)
		startDeferredWatchdog(workerCtx, cfg, logger, st, holder)
	}

	if cfg.NewAPIEnabled {
		go func() {
			n, err := registry.Deps.ModelsSvc.SyncPricingFromUpstream(context.Background())
			if err != nil {
				slog.Warn("startup pricing sync failed", "error", err)
			} else if n > 0 {
				slog.Info("startup pricing sync complete", "updated", n)
			}
		}()
	}

	return &App{
		Config:  cfg,
		Store:   st,
		Router:  router,
		Workers: bgWorkers,
		closers: []func(){
			cancel,
			func() { bgWorkers.stop(context.Background()) },
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
