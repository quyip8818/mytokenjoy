package handler

import (
	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	audithandler "github.com/tokenjoy/backend/internal/http/handler/audit"
	"github.com/tokenjoy/backend/internal/http/handler/auth"
	"github.com/tokenjoy/backend/internal/http/handler/billing"
	budgethandler "github.com/tokenjoy/backend/internal/http/handler/budget"
	dashboardhandler "github.com/tokenjoy/backend/internal/http/handler/dashboard"
	keyshandler "github.com/tokenjoy/backend/internal/http/handler/keys"
	mehandler "github.com/tokenjoy/backend/internal/http/handler/me"
	modelshandler "github.com/tokenjoy/backend/internal/http/handler/models"
	orghandler "github.com/tokenjoy/backend/internal/http/handler/org"
	"github.com/tokenjoy/backend/internal/http/handler/platform"
)

type Registry struct {
	cfg       httpdeps.Deps
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
	me        *mehandler.Handler
	internalIngest *InternalIngestHandler
}

func NewRegistry(deps httpdeps.Deps) Registry {
	p := deps.Protected()
	return Registry{
		cfg:       deps,
		session:   NewSessionHandler(p),
		auth:      auth.NewHandler(deps.Public(), deps.CompanySvc),
		platform:  platform.NewHandler(deps.Platform()),
		billing:   billing.NewHandler(p, deps.BillingSvc),
		org:       orghandler.NewHandler(p, deps.OrgSvc),
		budget:    budgethandler.NewHandler(p, deps.BudgetSvc),
		keys:      keyshandler.NewHandler(p, deps.KeysSvc),
		models:    modelshandler.NewHandler(p, deps.ModelsSvc),
		dashboard: dashboardhandler.NewHandler(p, deps.DashboardSvc),
		audit:     audithandler.NewHandler(p, deps.AuditSvc, deps.ReadModel),
		me:        mehandler.NewHandler(p, deps.MemberSvc),
		internalIngest: NewInternalIngestHandler(deps.Config, deps.IngestSvc, deps.IngestFailureRecorder, deps.IngestMetrics, deps.Logger),
	}
}

func (reg Registry) RegisterAPIRoutes(r chi.Router) {
	reg.session.RegisterRoutes(r)
	reg.auth.RegisterRoutes(r)
	r.Route("/internal", func(r chi.Router) {
		reg.internalIngest.RegisterRoutes(r)
	})
	reg.billing.RegisterRoutes(r)
	if reg.cfg.Config.SupportSaas {
		r.Route("/platform", reg.platform.RegisterRoutes)
	}
	r.Route("/org", reg.org.RegisterRoutes)
	r.Route("/budget", reg.budget.RegisterRoutes)
	r.Route("/keys", reg.keys.RegisterRoutes)
	r.Route("/models", reg.models.RegisterRoutes)
	r.Route("/dashboard", reg.dashboard.RegisterRoutes)
	r.Route("/audit", reg.audit.RegisterRoutes)
	r.Route("/me", reg.me.RegisterRoutes)
}
