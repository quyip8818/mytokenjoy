package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	audithandler "github.com/tokenjoy/backend/internal/http/handler/audit"
	"github.com/tokenjoy/backend/internal/http/handler/auth"
	"github.com/tokenjoy/backend/internal/http/handler/billing"
	budgethandler "github.com/tokenjoy/backend/internal/http/handler/budget"
	dashboardhandler "github.com/tokenjoy/backend/internal/http/handler/dashboard"
	keyshandler "github.com/tokenjoy/backend/internal/http/handler/keys"
	modelshandler "github.com/tokenjoy/backend/internal/http/handler/models"
	orghandler "github.com/tokenjoy/backend/internal/http/handler/org"
	"github.com/tokenjoy/backend/internal/http/handler/platform"
)

type Registry struct {
	cfg       config.Config
	session   *SessionHandler
	auth      *auth.Handler
	platform  *platform.Handler
	billing   *billing.Handler
	org       *orghandler.Handler
	budget    *budgethandler.Handler
	keys      *keyshandler.Handler
	models    *modelshandler.Handler
	dashboard *dashboardhandler.Handler
	audit     *audithandler.Handler
	webhook   *WebhookHandler
}

func NewRegistry(deps httpdeps.Deps) Registry {
	return Registry{
		cfg:       deps.Config,
		session:   NewSessionHandler(deps.Config, deps.SessionSvc),
		auth:      auth.NewHandler(deps.CompanySvc),
		platform:  platform.NewHandler(deps.Config, deps.CompanySvc, deps.BillingSvc, deps.KeysSvc, deps.PlatformSvc),
		billing:   billing.NewHandler(deps.Config, deps.BillingSvc, deps.SessionSvc),
		org:       orghandler.NewHandler(deps.Config, deps.OrgSvc, deps.SessionSvc),
		budget:    budgethandler.NewHandler(deps.Config, deps.BudgetSvc, deps.SessionSvc),
		keys:      keyshandler.NewHandler(deps.Config, deps.KeysSvc, deps.SessionSvc),
		models:    modelshandler.NewHandler(deps.Config, deps.ModelsSvc, deps.SessionSvc),
		dashboard: dashboardhandler.NewHandler(deps.Config, deps.DashboardSvc, deps.SessionSvc),
		audit:     audithandler.NewHandler(deps.Config, deps.AuditSvc, deps.ReadModel, deps.SessionSvc),
		webhook:   NewWebhookHandler(deps.Config, deps.IngestSvc, deps.Logger),
	}
}

func (reg Registry) RegisterAPIRoutes(r chi.Router) {
	reg.session.RegisterRoutes(r)
	reg.auth.RegisterRoutes(r)
	reg.webhook.RegisterRoutes(r)
	reg.billing.RegisterRoutes(r)
	if reg.cfg.SupportSaas {
		r.Route("/platform", reg.platform.RegisterRoutes)
	}
	r.Route("/org", reg.org.RegisterRoutes)
	r.Route("/budget", reg.budget.RegisterRoutes)
	r.Route("/keys", reg.keys.RegisterRoutes)
	r.Route("/models", reg.models.RegisterRoutes)
	r.Route("/dashboard", reg.dashboard.RegisterRoutes)
	r.Route("/audit", reg.audit.RegisterRoutes)
}
