package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/session"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/notification"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/internal/worker"
)

func openStore(ctx context.Context, cfg config.Config) (store.Store, error) {
	snapshot := seed.Load(cfg)
	if cfg.DatabaseURL == "" {
		return store.NewMemory(snapshot), nil
	}
	return postgres.New(ctx, cfg.DatabaseURL, snapshot)
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
	st, err := openStore(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	var adminClient newapi.AdminClient
	if cfg.NewAPIEnabled {
		adminClient = newapi.NewClient(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken)
	}

	lifecycle := relay.NewTokenLifecycle(cfg, st, adminClient)
	notifier := notification.NewService(cfg, st, logger)
	ingest := domainbudget.NewIngestService(cfg, st, lifecycle, notifier, logger)
	rebalance := domainbudget.NewRebalanceService(cfg, st, adminClient, lifecycle)

	sessionSvc := session.NewService(st)
	factory := datasource.NewFactory(cfg)
	orgSvc := domainorg.NewService(cfg, st, factory, lifecycle, notifier, logger)
	budgetSvc := domainbudget.NewService(cfg, st)
	keysSvc := domainkeys.NewService(cfg, st, lifecycle)
	modelsSvc := domainmodels.NewService(cfg, st, adminClient, lifecycle)
	logAggregator := domainusage.NewLogAggregator(adminClient, st, logger)
	dashboardSvc := domaindashboard.NewService(cfg, st, logAggregator)
	auditSvc := domainaudit.NewService(cfg, st)

	runner := worker.NewRunner(cfg, st, adminClient, lifecycle, ingest, rebalance, orgSvc, logger)

	router := httpapi.NewRouter(httpapi.Deps{
		Config:       cfg,
		Logger:       logger,
		SessionSvc:   sessionSvc,
		OrgSvc:       orgSvc,
		BudgetSvc:    budgetSvc,
		KeysSvc:      keysSvc,
		ModelsSvc:    modelsSvc,
		DashboardSvc: dashboardSvc,
		AuditSvc:     auditSvc,
		IngestSvc:    ingest,
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
