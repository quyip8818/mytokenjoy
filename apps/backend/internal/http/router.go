package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/session"
	httphandler "github.com/tokenjoy/backend/internal/http/handler"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
)

type Deps struct {
	Config       config.Config
	Logger       *slog.Logger
	SessionSvc   session.Service
	OrgSvc       domainorg.Service
	BudgetSvc    domainbudget.Service
	KeysSvc      domainkeys.Service
	ModelsSvc    domainmodels.Service
	DashboardSvc domaindashboard.Service
	AuditSvc     domainaudit.Service
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.RequestID)
	r.Use(httpmiddleware.Recover(deps.Logger))
	r.Use(httpmiddleware.CORS(deps.Config.CORSOriginList()))

	sessionHandler := httphandler.NewSessionHandler(deps.SessionSvc)
	orgHandler := httphandler.NewOrgHandler(deps.OrgSvc)
	budgetHandler := httphandler.NewBudgetHandler(deps.BudgetSvc)
	keysHandler := httphandler.NewKeysHandler(deps.KeysSvc)
	modelsHandler := httphandler.NewModelsHandler(deps.ModelsSvc)
	dashboardHandler := httphandler.NewDashboardHandler(deps.DashboardSvc)
	auditHandler := httphandler.NewAuditHandler(deps.AuditSvc)

	httphandler.RegisterHealthRoutes(r)

	r.Route("/api", func(r chi.Router) {
		sessionHandler.RegisterRoutes(r)
		r.Route("/org", orgHandler.RegisterRoutes)
		r.Route("/budget", budgetHandler.RegisterRoutes)
		r.Route("/keys", keysHandler.RegisterRoutes)
		r.Route("/models", modelsHandler.RegisterRoutes)
		r.Route("/dashboard", dashboardHandler.RegisterRoutes)
		r.Route("/audit", auditHandler.RegisterRoutes)
	})

	return r
}
