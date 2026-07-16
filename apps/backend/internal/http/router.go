package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	httphandler "github.com/tokenjoy/backend/internal/http/handler"
	healthhandler "github.com/tokenjoy/backend/internal/http/handler/health"
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
	reg := httphandler.NewRegistry(deps)

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
