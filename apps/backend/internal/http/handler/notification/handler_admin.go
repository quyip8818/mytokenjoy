package notification

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/store"
)

// --- Admin: Log ---

type adminLogEntryResponse struct {
	ID        uuid.UUID `json:"id"`
	Channel   string    `json:"channel"`
	EventType string    `json:"eventType"`
	UserID    uuid.UUID `json:"userId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	CreatedAt string    `json:"createdAt"`
	ReadAt    *string   `json:"readAt,omitempty"`
}

func (h *Handler) AdminLog(w http.ResponseWriter, r *http.Request) {
	filter := types.NotificationLogFilter{
		Channel:   r.URL.Query().Get("channel"),
		Status:    r.URL.Query().Get("status"),
		EventType: r.URL.Query().Get("eventType"),
	}
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))

	entries, err := h.store.Notification().ListLog(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	items := make([]adminLogEntryResponse, len(entries))
	for i, e := range entries {
		items[i] = adminLogEntryResponse{
			ID:        e.ID,
			Channel:   e.Channel,
			EventType: e.EventType,
			UserID:    e.UserID,
			Title:     e.Title,
			Body:      e.Body,
			Status:    e.Status,
			Error:     e.Error,
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
		}
		if e.ReadAt != nil {
			s := e.ReadAt.Format(time.RFC3339)
			items[i].ReadAt = &s
		}
	}

	httputil.WriteOK(w, items)
}

// --- Admin: Stats ---

type adminStatResponse struct {
	Channel string `json:"channel"`
	Status  string `json:"status"`
	Count   int    `json:"count"`
}

func (h *Handler) AdminStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.Notification().Stats(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	items := make([]adminStatResponse, len(stats))
	for i, s := range stats {
		items[i] = adminStatResponse{
			Channel: s.Channel,
			Status:  s.Status,
			Count:   s.Count,
		}
	}

	httputil.WriteOK(w, items)
}

// --- Admin: Test Send ---

type adminTestSendRequest struct {
	UserID    uuid.UUID `json:"userId"`
	EventType string    `json:"eventType"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
}

func (h *Handler) AdminTestSend(w http.ResponseWriter, r *http.Request) {
	var req adminTestSendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == uuid.Nil || req.Title == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "userId and title required")
		return
	}

	eventType := req.EventType
	if eventType == "" {
		eventType = "admin_test"
	}

	event := domainnotification.Event{
		EventType:   eventType,
		RecipientID: req.UserID,
		CompanyID:   store.CompanyID(r.Context()),
		Payload: map[string]any{
			"title": req.Title,
			"body":  req.Body,
		},
		Metadata: domainnotification.EventMetadata{
			Priority: domainnotification.PriorityNormal,
		},
	}

	err := h.notifySvc.Dispatch(r.Context(), event)
	httputil.WriteVoid(w, err)
}
