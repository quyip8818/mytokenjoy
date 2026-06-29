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
	"github.com/tokenjoy/backend/internal/http/response"
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
	IngestSvc    domainbudget.Ingestor
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.NotFound(jsonNotFound)
	r.MethodNotAllowed(jsonMethodNotAllowed)
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.RequestID)
	r.Use(httpmiddleware.Recover(deps.Logger))
	r.Use(httpmiddleware.CORS(deps.Config.CORSOriginList()))

	reg := httphandler.NewRegistry(httphandler.RegistryDeps(deps))

	httphandler.RegisterHealthRoutes(r)

	r.Route("/api", reg.RegisterAPIRoutes)

	return r
}

func jsonNotFound(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotFound, "Not found")
}

func jsonMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
}
