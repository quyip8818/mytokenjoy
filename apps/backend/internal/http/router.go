package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	httphandler "github.com/tokenjoy/backend/internal/http/handler"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/http/response"
)

func NewRouter(deps httpdeps.Deps) http.Handler {
	r := chi.NewRouter()
	r.NotFound(jsonNotFound)
	r.MethodNotAllowed(jsonMethodNotAllowed)
	r.Use(middleware.RealIP)
	r.Use(httpmiddleware.RequestID)
	r.Use(httpmiddleware.Recover(deps.Logger))
	r.Use(httpmiddleware.CORS(deps.Config.CORSOriginList()))

	if deps.Config.RelayGatewayEnabled && deps.Config.NewAPIEnabled && deps.RelayGateway != nil {
		r.Handle("/v1/*", deps.RelayGateway)
	} else if deps.Config.RelayGatewayEnabled && deps.Config.NewAPIEnabled && deps.Logger != nil {
		deps.Logger.Error("relay gateway disabled", "error", "gateway service unavailable")
	}

	reg := httphandler.NewRegistry(deps)

	httphandler.RegisterHealthRoutes(r)

	r.Route("/api", func(api chi.Router) {
		api.Use(httpmiddleware.CompanyResolve(deps.Config, deps.CompanySvc, deps.SessionToken))
		api.Use(httpmiddleware.AuthzRevisionHeader(deps.Store))
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
