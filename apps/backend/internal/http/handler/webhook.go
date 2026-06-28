package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

type WebhookHandler struct {
	cfg    config.Config
	ingest *domainbudget.IngestService
}

func NewWebhookHandler(cfg config.Config, ingest *domainbudget.IngestService) *WebhookHandler {
	return &WebhookHandler{cfg: cfg, ingest: ingest}
}

func (h *WebhookHandler) RegisterRoutes(r chi.Router) {
	r.Route("/internal", func(r chi.Router) {
		r.Post("/webhooks/newapi-log", h.HandleNewAPILog)
	})
}

func (h *WebhookHandler) HandleNewAPILog(w http.ResponseWriter, r *http.Request) {
	if h.cfg.NewAPIWebhookSecret == "" || r.Header.Get("X-Webhook-Secret") != h.cfg.NewAPIWebhookSecret {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var payload newapi.WebhookLogPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	if err := newapi.ValidateWebhookPayload(payload); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.ingest.Ingest(r.Context(), payload); err != nil {
		_ = h.ingest.EnqueueFailed(payload, err)
		response.Error(w, http.StatusInternalServerError, "Ingest failed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
