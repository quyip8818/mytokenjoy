package app

import (
	"context"
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
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/integration/newapi"
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

func New(cfg config.Config, logger *slog.Logger) *App {
	ctx := context.Background()
	st, err := openStore(ctx, cfg)
	if err != nil {
		logger.Error("open store", "error", err)
		panic(err)
	}

	var adminClient newapi.AdminClient
	if cfg.NewAPIEnabled {
		adminClient = newapi.NewClient(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken)
	}

	lifecycle := relay.NewTokenLifecycle(cfg, st, adminClient)
	ingest := domainbudget.NewIngestService(cfg, st, lifecycle, logger)
	rebalance := domainbudget.NewRebalanceService(cfg, st, adminClient, lifecycle)

	sessionSvc := session.NewService(st)
	orgSvc := domainorg.NewService(cfg, st)
	budgetSvc := domainbudget.NewService(cfg, st)
	keysSvc := domainkeys.NewService(cfg, st, lifecycle)
	modelsSvc := domainmodels.NewService(cfg, st, adminClient, lifecycle)
	dashboardSvc := domaindashboard.NewService(cfg, st)
	auditSvc := domainaudit.NewService(cfg, st)

	runner := worker.NewRunner(cfg, st, adminClient, lifecycle, ingest, rebalance, logger)

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
	runner.Start(workerCtx)

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
	}
}

func (a *App) Close() {
	for _, closer := range a.closers {
		closer()
	}
}
