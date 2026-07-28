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
	r.Get("/versions", h.Versions)
	r.Get("/catalog/models", h.CatalogModels)
	r.Get("/catalog/channels", h.CatalogChannels)
	// Legacy: full catalog (backward compat)
	r.Get("/catalog", h.Catalog)
}

// Versions returns the current version number of each sync partition.
func (h *Handler) Versions(w http.ResponseWriter, r *http.Request) {
	versions, err := h.service.GetVersions(r.Context())
	if err != nil {
		http.Error(w, `{"message":"failed to load versions"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

// CatalogModels returns the models partition with version + data.
func (h *Handler) CatalogModels(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetModels(r.Context())
	if err != nil {
		http.Error(w, `{"message":"failed to load models catalog"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CatalogChannels returns the channels partition with version + data.
func (h *Handler) CatalogChannels(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.GetChannels(r.Context())
	if err != nil {
		http.Error(w, `{"message":"failed to load channels catalog"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Catalog is the legacy full-catalog endpoint (backward compat).
func (h *Handler) Catalog(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.service.GetCatalog(r.Context())
	if err != nil {
		http.Error(w, `{"message":"failed to load catalog"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
}
