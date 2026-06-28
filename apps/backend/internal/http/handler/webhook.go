package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

type WebhookHandler struct {
	cfg    config.Config
	ingest domainbudget.Ingestor
}

func NewWebhookHandler(cfg config.Config, ingest domainbudget.Ingestor) *WebhookHandler {
	return &WebhookHandler{cfg: cfg, ingest: ingest}
}

func (h *WebhookHandler) RegisterRoutes(r chi.Router) {
	r.Route("/internal", func(r chi.Router) {
		r.Post("/webhooks/newapi-log", h.HandleNewAPILog)
	})
}

func (h *WebhookHandler) HandleNewAPILog(w http.ResponseWriter, r *http.Request) {
	if h.cfg.NewAPIWebhookSecret == "" || r.Header.Get("X-Webhook-Secret") != h.cfg.NewAPIWebhookSecret {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	var payload newapi.WebhookLogPayload
	if err := httputil.DecodeJSON(r, &payload); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if err := newapi.ValidateWebhookPayload(payload); err != nil {
		httputil.WriteError(w, domain.BadRequest(err.Error()))
		return
	}
	if err := h.ingest.Ingest(r.Context(), payload); err != nil {
		_ = h.ingest.EnqueueFailed(payload, err)
		httputil.WriteStatus(w, http.StatusInternalServerError, "Ingest failed")
		return
	}
	httputil.WriteOK(w, map[string]string{"status": "ok"})
}
