package app

import (
	"log/slog"
	"net/http"

	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/session"
	httpapi "github.com/tokenjoy/backend/internal/http"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
)

type App struct {
	Config config.Config
	Store  store.Store
	Router http.Handler
}

func New(cfg config.Config, logger *slog.Logger) *App {
	st := store.NewMemory(seed.Load(cfg))
	sessionSvc := session.NewService(st)
	orgSvc := domainorg.NewService(cfg, st)
	budgetSvc := domainbudget.NewService(cfg, st)
	keysSvc := domainkeys.NewService(cfg, st)
	modelsSvc := domainmodels.NewService(cfg, st)
	dashboardSvc := domaindashboard.NewService(cfg, st)
	auditSvc := domainaudit.NewService(cfg, st)

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
	})

	return &App{
		Config: cfg,
		Store:  st,
		Router: router,
	}
}
