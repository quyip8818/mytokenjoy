package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/tokenjoy/backend/internal/adapter/enqueue"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/platform"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/internal/worker/pricingsync"
	"github.com/tokenjoy/backend/internal/worker/smssync"
	smsintegration "github.com/tokenjoy/backend/internal/integration/sms"
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
	skipWorker bool
	adminPort  adminport.Port
	orgSync    domainorg.SyncService
}

type Option func(*options)

func WithoutWorker() Option {
	return func(o *options) {
		o.skipWorker = true
	}
}

func WithAdminPort(client adminport.Port) Option {
	return func(o *options) {
		o.adminPort = client
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
	orgAdmin := enqueue.NewOrgRiverAdminHolder(nil)
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
	router := httpapi.NewRouter(registry.Deps)

	workerCtx, cancel := context.WithCancel(context.Background())
	if !o.skipWorker {
		bgWorkers.start(workerCtx, cfg)
		startDeferredWatchdog(workerCtx, cfg, logger, st, holder)
		startPricingSyncWorker(workerCtx, cfg, registry.Infra.adminPort)
		startSMSSyncWorker(workerCtx, cfg, registry.Infra.adminPort, st)
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

func startPricingSyncWorker(ctx context.Context, cfg config.Config, adminPort adminport.Port) {
	if !cfg.PlatformPricingSyncEnabled {
		return
	}
	if cfg.PlatformPricingSyncURL == "" || cfg.PlatformPricingSyncKey == "" {
		slog.Warn("pricing sync enabled but URL/Key not configured, skipping")
		return
	}
	interval := time.Duration(cfg.PlatformPricingSyncIntervalSec) * time.Second
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	pc := platform.NewClient(cfg.PlatformPricingSyncURL, cfg.PlatformPricingSyncKey)
	w := pricingsync.New(pc, adminPort, interval)
	go w.Run(ctx)
	slog.Info("pricing sync worker started", "interval", interval)
}

func startSMSSyncWorker(ctx context.Context, cfg config.Config, adminPort adminport.Port, st store.Store) {
	if !cfg.SMSSyncEnabled {
		return
	}
	if cfg.SMSAPIBaseURL == "" || cfg.SMSClientID == "" || cfg.SMSClientSecret == "" {
		slog.Warn("sms sync enabled but config incomplete, skipping")
		return
	}
	interval := time.Duration(cfg.SMSSyncIntervalSec) * time.Second
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	client := smsintegration.NewClient(smsintegration.Config{
		BaseURL:      cfg.SMSAPIBaseURL,
		ClientID:     cfg.SMSClientID,
		ClientSecret: cfg.SMSClientSecret,
	})
	// ponytail: ModelStore now uses real repo — writes to models table with source="sms"
	target := smssync.NewAdminPortTarget(adminPort, smssync.NewRepoModelStore(st.Models()))
	w := smssync.NewWithInterval(client, target, interval)
	go w.Run(ctx)
	slog.Info("sms sync worker started", "interval", interval, "url", cfg.SMSAPIBaseURL)
}

