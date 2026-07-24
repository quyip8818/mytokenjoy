package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	approvalhandler "github.com/tokenjoy/backend/internal/http/handler/approval"
	audithandler "github.com/tokenjoy/backend/internal/http/handler/audit"
	authhandler "github.com/tokenjoy/backend/internal/http/handler/auth"
	billinghandler "github.com/tokenjoy/backend/internal/http/handler/billing"
	budgethandler "github.com/tokenjoy/backend/internal/http/handler/budget"
	dashboardhandler "github.com/tokenjoy/backend/internal/http/handler/dashboard"
	devhandler "github.com/tokenjoy/backend/internal/http/handler/dev"
	healthhandler "github.com/tokenjoy/backend/internal/http/handler/health"
	ingesthandler "github.com/tokenjoy/backend/internal/http/handler/ingest"
	keyshandler "github.com/tokenjoy/backend/internal/http/handler/keys"
	mehandler "github.com/tokenjoy/backend/internal/http/handler/me"
	modelshandler "github.com/tokenjoy/backend/internal/http/handler/models"
	notificationhandler "github.com/tokenjoy/backend/internal/http/handler/notification"
	orghandler "github.com/tokenjoy/backend/internal/http/handler/org"
	platformhandler "github.com/tokenjoy/backend/internal/http/handler/platform"
	registerhandler "github.com/tokenjoy/backend/internal/http/handler/register"
	sessionhandler "github.com/tokenjoy/backend/internal/http/handler/session"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/infra/ratelimit"
)

func NewRouter(deps httpdeps.Deps) http.Handler {
	r := chi.NewRouter()
	r.NotFound(jsonNotFound)
	r.MethodNotAllowed(jsonMethodNotAllowed)

	// --- Global middleware (all routes) ---
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.RequestID)
	r.Use(httpmiddleware.LoggerContext(deps.Logger))
	r.Use(httpmiddleware.AccessLog(deps.Logger, deps.Config.AccessLogSlowMs))
	r.Use(httpmiddleware.Recover(deps.Logger))
	r.Use(httpmiddleware.SecurityHeaders(deps.Config.SecureCookie))
	r.Use(httpmiddleware.CORS(deps.Config.CORSOriginList()))

	// --- /v1 gateway (no timeout — streaming can last minutes) ---
	if deps.Config.GatewayEnabled && deps.Config.NewAPIEnabled && deps.Gateway != nil {
		r.Handle("/v1/*", deps.Gateway)
	} else if deps.Config.GatewayEnabled && deps.Config.NewAPIEnabled && deps.Logger != nil {
		deps.Logger.Error("newapi gateway disabled", "error", "gateway service unavailable")
	}

	// --- Health check ---
	healthhandler.RegisterRoutes(r)

	// --- /api routes ---
	r.Route("/api", func(api chi.Router) {
		api.Use(httpmiddleware.RequestTimeout(deps.Config.RequestTimeoutSec))
		api.Use(httpmiddleware.CompanyResolve(deps.Config, deps.CompanySvc, deps.SessionToken))
		if deps.Config.RateLimitEnabled {
			api.Use(httpmiddleware.RateLimitTenant(
				deps.RateLimiter,
				deps.Config.RateLimitTenantRate, deps.Config.RateLimitTenantBurst,
				deps.Config.RateLimitDryRun, deps.Logger,
			))
			api.Use(httpmiddleware.RateLimitLoginPaths(
				[]string{"/api/auth/login", "/api/auth/accept-invite", "/api/platform/auth/login"},
				deps.RateLimiter, ratelimit.NewMemoryLimiter(),
				deps.Config.RateLimitLoginMax, deps.Config.RateLimitLoginWindowSec,
				deps.Config.RateLimitDryRun, deps.Logger,
			))
		}
		api.Use(httpmiddleware.AuthzRevisionHeader(deps.AuthzSvc))
		if deps.CompanyGate != nil {
			api.Use(httpmiddleware.CompanyReadOnlyMiddleware(deps.CompanyGate))
		}
		mountAPI(api, deps)
	})

	return r
}

// mountAPI registers all API handlers. Adding a new handler = one line here.
func mountAPI(api chi.Router, d httpdeps.Deps) {
	sessionhandler.Mount(api, d)
	authhandler.Mount(api, d)
	ingesthandler.Mount(api, d)
	billinghandler.Mount(api, d)

	// SaaS only: register endpoints
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireSaaS(d.Config))
		registerhandler.Mount(r, d)
	})

	// SaaS only: platform management
	if d.Config.SupportSaas {
		platformhandler.Mount(api, d)
	}

	orghandler.Mount(api, d)
	budgethandler.Mount(api, d)
	keyshandler.Mount(api, d)
	modelshandler.Mount(api, d)
	dashboardhandler.Mount(api, d)
	audithandler.Mount(api, d)
	approvalhandler.Mount(api, d)
	mehandler.Mount(api, d)
	notificationhandler.Mount(api, d)

	if d.Config.AllowsDevHTTPRoutes() {
		devhandler.Mount(api, d)
	}
}

func jsonNotFound(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotFound, "Not found")
}

func jsonMethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
}
