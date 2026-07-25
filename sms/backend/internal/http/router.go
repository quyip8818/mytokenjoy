package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"sms/backend/internal/http/deps"
	"sms/backend/internal/http/handler"
	httpmiddleware "sms/backend/internal/http/middleware"
	"sms/backend/internal/http/response"
)

func NewRouter(d deps.Deps) http.Handler {
	r := chi.NewRouter()
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		response.Error(w, http.StatusNotFound, "Not found")
	})

	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(httpmiddleware.Logger(d.Logger))
	r.Use(httpmiddleware.CORS(d.Config.CORSOriginList()))

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	reg := handler.NewRegistry(d)
	r.Route("/api", func(api chi.Router) {
		reg.RegisterAPIRoutes(api)
	})

	return r
}
