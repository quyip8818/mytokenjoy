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
	budget    *BudgetHandler
	keys      *KeysHandler
	models    *ModelsHandler
	dashboard *DashboardHandler
	audit     *AuditHandler
	webhook   *WebhookHandler
}

func NewRegistry(deps RegistryDeps) Registry {
	return Registry{
		session:   NewSessionHandler(deps.Config, deps.SessionSvc),
		org:       orghandler.NewHandler(deps.Config, deps.OrgSvc, deps.SessionSvc),
		budget:    NewBudgetHandler(deps.Config, deps.BudgetSvc, deps.SessionSvc),
		keys:      NewKeysHandler(deps.Config, deps.KeysSvc, deps.SessionSvc),
		models:    NewModelsHandler(deps.Config, deps.ModelsSvc, deps.SessionSvc),
		dashboard: NewDashboardHandler(deps.Config, deps.DashboardSvc, deps.SessionSvc),
		audit:     NewAuditHandler(deps.Config, deps.AuditSvc, deps.SessionSvc),
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
