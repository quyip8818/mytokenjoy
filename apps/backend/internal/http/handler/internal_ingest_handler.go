package handler

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
)

type newAPILogWebhookRequest struct {
	LogID int64 `json:"log_id"`
}

type InternalIngestHandler struct {
	cfg     config.Config
	enqueue domainusage.Enqueuer
	metrics ingestmetrics.Recorder
	logger  *slog.Logger
}

func NewInternalIngestHandler(
	cfg config.Config,
	enqueue domainusage.Enqueuer,
	metrics ingestmetrics.Recorder,
	logger *slog.Logger,
) *InternalIngestHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if enqueue == nil {
		enqueue = domainusage.NewQueue(nil)
	}
	if metrics == nil {
		metrics = ingestmetrics.NoopCollector()
	}
	return &InternalIngestHandler{
		cfg:     cfg,
		enqueue: enqueue,
		metrics: metrics,
		logger:  logger,
	}
}

func (h *InternalIngestHandler) RegisterRoutes(r chi.Router) {
	r.Post("/webhooks/newapi-log", h.HandleNewAPILog)
	r.Get("/metrics/ingest", h.HandleIngestMetrics)
}

func (h *InternalIngestHandler) authenticateWebhookSecret(r *http.Request) bool {
	secret := r.Header.Get("X-Webhook-Secret")
	if h.cfg.NewAPIWebhookSecret == "" || secret == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(secret), []byte(h.cfg.NewAPIWebhookSecret)) == 1
}

func (h *InternalIngestHandler) HandleNewAPILog(w http.ResponseWriter, r *http.Request) {
	if !h.authenticateWebhookSecret(r) {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	var payload newAPILogWebhookRequest
	if err := httputil.DecodeJSON(r, &payload); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if payload.LogID <= 0 {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid log_id")
		return
	}

	if err := h.enqueue.Enqueue(r.Context(), payload.LogID, types.SourceWebhook); err != nil {
		h.logger.Error("enqueue ingest job", "log_id", payload.LogID, "error", err)
		httputil.WriteStatus(w, http.StatusInternalServerError, "enqueue failed")
		return
	}
	h.metrics.RecordNotifySuccess()
	httputil.WriteOK(w, map[string]string{"status": "accepted"})
}

func (h *InternalIngestHandler) HandleIngestMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.IngestEnabled() {
		httputil.WriteStatus(w, http.StatusNotFound, "ingest not enabled")
		return
	}
	if !h.authenticateWebhookSecret(r) {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	httputil.WriteOK(w, h.metrics.Snapshot())
}
