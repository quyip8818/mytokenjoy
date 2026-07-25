package newapisync

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	syncsvc "sms/backend/internal/domain/newapisync"
	"sms/backend/internal/http/response"
)

type Handler struct {
	svc *syncsvc.Service
}

func NewHandler(svc *syncsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/status", h.Status)
	r.Get("/models", h.Models)
	r.Post("/sync", h.Sync)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	statuses, err := h.svc.GetStatus(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, statuses)
}

func (h *Handler) Models(w http.ResponseWriter, r *http.Request) {
	models, err := h.svc.ListModels(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, models)
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	count, err := h.svc.SyncAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]int{"synced": count})
}
