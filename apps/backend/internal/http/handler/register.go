package handler

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/session"
	audithandler "github.com/tokenjoy/backend/internal/http/handler/audit"
	budgethandler "github.com/tokenjoy/backend/internal/http/handler/budget"
	dashboardhandler "github.com/tokenjoy/backend/internal/http/handler/dashboard"
	keyshandler "github.com/tokenjoy/backend/internal/http/handler/keys"
	modelshandler "github.com/tokenjoy/backend/internal/http/handler/models"
	orghandler "github.com/tokenjoy/backend/internal/http/handler/org"
)

type RegistryDeps struct {
	Config       config.Config
	Logger       *slog.Logger
	SessionSvc   session.Service
	OrgSvc       domainorg.Service
	BudgetSvc    domainbudget.Service
	KeysSvc      domainkeys.Service
	ModelsSvc    domainmodels.Service
	DashboardSvc domaindashboard.Service
	AuditSvc     domainaudit.Service
	IngestSvc    domainbudget.Ingestor
}

type Registry struct {
	session   *SessionHandler
	org       *orghandler.Handler
	budget    *budgethandler.Handler
	keys      *keyshandler.Handler
	models    *modelshandler.Handler
	dashboard *dashboardhandler.Handler
	audit     *audithandler.Handler
	webhook   *WebhookHandler
}

func NewRegistry(deps RegistryDeps) Registry {
	return Registry{
		session:   NewSessionHandler(deps.Config, deps.SessionSvc),
		org:       orghandler.NewHandler(deps.Config, deps.OrgSvc, deps.SessionSvc),
		budget:    budgethandler.NewHandler(deps.Config, deps.BudgetSvc, deps.SessionSvc),
		keys:      keyshandler.NewHandler(deps.Config, deps.KeysSvc, deps.SessionSvc),
		models:    modelshandler.NewHandler(deps.Config, deps.ModelsSvc, deps.SessionSvc),
		dashboard: dashboardhandler.NewHandler(deps.Config, deps.DashboardSvc, deps.SessionSvc),
		audit:     audithandler.NewHandler(deps.Config, deps.AuditSvc, deps.SessionSvc),
		webhook:   NewWebhookHandler(deps.Config, deps.IngestSvc, deps.Logger),
	}
}

func (reg Registry) RegisterAPIRoutes(r chi.Router) {
	reg.session.RegisterRoutes(r)
	reg.webhook.RegisterRoutes(r)
	r.Route("/org", reg.org.RegisterRoutes)
	r.Route("/budget", reg.budget.RegisterRoutes)
	r.Route("/keys", reg.keys.RegisterRoutes)
	r.Route("/models", reg.models.RegisterRoutes)
	r.Route("/dashboard", reg.dashboard.RegisterRoutes)
	r.Route("/audit", reg.audit.RegisterRoutes)
}
