package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

type WebhookHandler struct {
	cfg    config.Config
	ingest domainusage.Ingestor
	logger *slog.Logger
}

func NewWebhookHandler(cfg config.Config, ingest domainusage.Ingestor, logger *slog.Logger) *WebhookHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &WebhookHandler{cfg: cfg, ingest: ingest, logger: logger}
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
	if err := h.ingest.Ingest(r.Context(), payload, types.SourceWebhook); err != nil {
		if enqueueErr := h.ingest.EnqueueFailed(r.Context(), payload, err); enqueueErr != nil {
			h.logger.Error("ingest failed and enqueue to outbox failed",
				"ingest_error", err,
				"enqueue_error", enqueueErr,
			)
		}
		httputil.WriteStatus(w, http.StatusInternalServerError, "Ingest failed")
		return
	}
	httputil.WriteOK(w, map[string]string{"status": "ok"})
}
