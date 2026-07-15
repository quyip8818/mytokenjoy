package handler

import (
	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	audithandler "github.com/tokenjoy/backend/internal/http/handler/audit"
	"github.com/tokenjoy/backend/internal/http/handler/auth"
	"github.com/tokenjoy/backend/internal/http/handler/billing"
	budgethandler "github.com/tokenjoy/backend/internal/http/handler/budget"
	dashboardhandler "github.com/tokenjoy/backend/internal/http/handler/dashboard"
	devhandler "github.com/tokenjoy/backend/internal/http/handler/dev"
	ingesthandler "github.com/tokenjoy/backend/internal/http/handler/ingest"
	keyshandler "github.com/tokenjoy/backend/internal/http/handler/keys"
	mehandler "github.com/tokenjoy/backend/internal/http/handler/me"
	modelshandler "github.com/tokenjoy/backend/internal/http/handler/models"
	notificationhandler "github.com/tokenjoy/backend/internal/http/handler/notification"
	orghandler "github.com/tokenjoy/backend/internal/http/handler/org"
	"github.com/tokenjoy/backend/internal/http/handler/platform"
	sessionhandler "github.com/tokenjoy/backend/internal/http/handler/session"
)

type Registry struct {
	cfg            httpdeps.Deps
	session        *sessionhandler.Handler
	auth           *auth.Handler
	platform       *platform.Handler
	billing        *billing.Handler
	org            *orghandler.Handler
	budget         *budgethandler.Handler
	keys           *keyshandler.Handler
	models         *modelshandler.Handler
	dashboard      *dashboardhandler.Handler
	audit          *audithandler.Handler
	me             *mehandler.Handler
	notification   *notificationhandler.Handler
	internalIngest *ingesthandler.Handler
	dev            *devhandler.Handler
}

func NewRegistry(deps httpdeps.Deps) Registry {
	p := deps.Protected()
	return Registry{
		cfg:            deps,
		session:        sessionhandler.NewHandler(p),
		auth:           auth.NewHandler(deps.Public(), deps.CompanySvc),
		platform:       platform.NewHandler(deps.Platform()),
		billing:        billing.NewHandler(p, deps.BillingSvc),
		org:            orghandler.NewHandler(p, deps.OrgSvc, deps.CompanySvc),
		budget:         budgethandler.NewHandler(p, deps.BudgetSvc),
		keys:           keyshandler.NewHandler(p, deps.KeysSvc),
		models:         modelshandler.NewHandler(p, deps.ModelsSvc),
		dashboard:      dashboardhandler.NewHandler(p, deps.DashboardSvc),
		audit:          audithandler.NewHandler(p, deps.AuditSvc),
		me:             mehandler.NewHandler(p, deps.MemberAnalyticsSvc),
		notification:   notificationhandler.NewHandler(p, deps.Store, deps.NotificationSvc),
		internalIngest: ingesthandler.NewHandler(deps.Config, deps.IngestQueue, deps.IngestMetrics, deps.Logger),
		dev:            devhandler.NewHandler(p, deps.Config.LocalCompanyID, deps.DevBearerResolver, deps.DevReadinessChecker),
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
	r.Route("/notifications", reg.notification.RegisterRoutes)
	// /api/dev/* — local development only; see config.AllowsDevHTTPRoutes.
	if reg.cfg.Config.AllowsDevHTTPRoutes() {
		r.Route("/dev", reg.dev.RegisterRoutes)
	}
}
