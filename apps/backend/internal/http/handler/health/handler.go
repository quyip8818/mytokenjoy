package health

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/http/response"
)

func RegisterRoutes(r chi.Router) {
	r.Get("/healthz", healthz)
	r.Head("/healthz", healthzHead)
	r.Post("/healthz", healthz)
	r.Delete("/healthz", healthz)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func healthzHead(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
