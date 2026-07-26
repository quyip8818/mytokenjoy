package sync

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	syncsvc "sms/backend/internal/domain/sync"
)

type Handler struct {
	service *syncsvc.Service
}

func NewHandler(service *syncsvc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/catalog", h.Catalog)
}

func (h *Handler) Catalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.service.GetCatalog(r.Context())
	if err != nil {
		http.Error(w, `{"message":"failed to load catalog"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
}
