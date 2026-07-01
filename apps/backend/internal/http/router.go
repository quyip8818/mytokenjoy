package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/session"
	httphandler "github.com/tokenjoy/backend/internal/http/handler"
	relayhandler "github.com/tokenjoy/backend/internal/http/handler/relay"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/store"
)

type Deps struct {
	Config       config.Config
	Logger       *slog.Logger
	Store        store.Store
	SessionSvc   session.Service
	OrgSvc       domainorg.Service
	BudgetSvc    domainbudget.Service
	KeysSvc      domainkeys.Service
	ModelsSvc    domainmodels.Service
	DashboardSvc domaindashboard.Service
	AuditSvc     domainaudit.Service
	IngestSvc    domainbudget.Ingestor
	CompanySvc   domaincompany.Service
	BillingSvc   domainbilling.Service
	PlatformSvc  httpmiddleware.PlatformService
	WalletSvc    domaincompany.WalletService
	CompanyGate  *domaincompany.Gate
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.NotFound(jsonNotFound)
	r.MethodNotAllowed(jsonMethodNotAllowed)
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.RequestID)
	r.Use(httpmiddleware.Recover(deps.Logger))
	r.Use(httpmiddleware.CORS(deps.Config.CORSOriginList()))

	if deps.Config.RelayGatewayEnabled && deps.Config.NewAPIEnabled {
		if gw, err := relayhandler.NewGateway(deps.Config, deps.Store, deps.WalletSvc); err == nil {
			r.Handle("/v1/*", gw)
		} else if deps.Logger != nil {
			deps.Logger.Error("relay gateway disabled", "error", err)
		}
	}

	reg := httphandler.NewRegistry(httphandler.RegistryDeps{
		Config:       deps.Config,
		Logger:       deps.Logger,
		SessionSvc:   deps.SessionSvc,
		OrgSvc:       deps.OrgSvc,
		BudgetSvc:    deps.BudgetSvc,
		KeysSvc:      deps.KeysSvc,
		ModelsSvc:    deps.ModelsSvc,
		DashboardSvc: deps.DashboardSvc,
		AuditSvc:     deps.AuditSvc,
		IngestSvc:    deps.IngestSvc,
		CompanySvc:   deps.CompanySvc,
		BillingSvc:   deps.BillingSvc,
		PlatformSvc:  deps.PlatformSvc,
	})

	httphandler.RegisterHealthRoutes(r)

	r.Route("/api", func(api chi.Router) {
		api.Use(httpmiddleware.CompanyResolve(deps.Config, deps.CompanySvc))
		if deps.CompanyGate != nil {
			api.Use(domaincompany.CompanyReadOnlyMiddleware(deps.CompanyGate))
		}
		reg.RegisterAPIRoutes(api)
	})

	return r
}

func jsonNotFound(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotFound, "Not found")
}

func jsonMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
}
