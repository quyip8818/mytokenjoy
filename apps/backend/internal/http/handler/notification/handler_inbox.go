package notification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
)

// --- List notifications (cursor pagination + filters + grouping) ---

type notificationItemResponse struct {
	ID         uuid.UUID       `json:"id"`
	EventType  string          `json:"eventType"`
	Channel    string          `json:"channel"`
	Category   string          `json:"category"`
	Title      string          `json:"title"`
	Body       string          `json:"body"`
	Payload    json.RawMessage `json:"payload"`
	GroupKey   string          `json:"groupKey,omitempty"`
	GroupCount int             `json:"groupCount,omitempty"`
	Status     string          `json:"status"`
	CreatedAt  string          `json:"createdAt"`
	ReadAt     *string         `json:"readAt"`
	UpdatedAt  string          `json:"updatedAt"`
}

type listResponse struct {
	Items      []notificationItemResponse `json:"items"`
	NextCursor *string                    `json:"nextCursor"`
	HasMore    bool                       `json:"hasMore"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	filter := types.NotificationListFilter{
		UserID:   sessionCtx.Member.ID,
		Category: q.Get("category"),
		Status:   q.Get("status"),
		Archived: q.Get("archived") == "true",
		Grouped:  q.Get("grouped") != "false", // default true
		GroupKey: q.Get("group_key"),
		Limit:    limit,
	}

	if cursor := q.Get("cursor"); cursor != "" {
		t, err := time.Parse(time.RFC3339Nano, cursor)
		if err == nil {
			filter.Cursor = &t
		}
	}

	result, err := h.notifRepo.List(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	items := make([]notificationItemResponse, len(result.Items))
	for i, e := range result.Items {
		items[i] = notificationItemResponse{
			ID:         e.ID,
			EventType:  e.EventType,
			Channel:    e.Channel,
			Category:   e.Category,
			Title:      e.Title,
			Body:       e.Body,
			Payload:    e.Payload,
			GroupKey:   e.GroupKey,
			GroupCount: e.GroupCount,
			Status:     e.Status,
			CreatedAt:  e.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  e.UpdatedAt.Format(time.RFC3339),
		}
		if e.ReadAt != nil {
			s := e.ReadAt.Format(time.RFC3339)
			items[i].ReadAt = &s
		}
	}

	var nextCursor *string
	if result.NextCursor != nil {
		s := result.NextCursor.UTC().Format(time.RFC3339Nano)
		nextCursor = &s
	}

	httputil.WriteOK(w, listResponse{Items: items, NextCursor: nextCursor, HasMore: result.HasMore})
}

// --- Unread count ---

type unreadCountResponse struct {
	Count int `json:"count"`
}

func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	count, err := h.notifRepo.GetUnreadCount(r.Context(), sessionCtx.Member.ID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	httputil.WriteOK(w, unreadCountResponse{Count: count})
}

// --- Mark read ---

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	_, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.notifRepo.MarkRead(r.Context(), id)
	httputil.WriteVoid(w, err)
}

// --- Mark all read ---

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	err := h.notifRepo.MarkAllRead(r.Context(), sessionCtx.Member.ID)
	httputil.WriteVoid(w, err)
}

// --- Archive ---

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	_, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.notifRepo.Archive(r.Context(), id)
	httputil.WriteVoid(w, err)
}

// --- Unarchive ---

func (h *Handler) Unarchive(w http.ResponseWriter, r *http.Request) {
	_, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.notifRepo.Unarchive(r.Context(), id)
	httputil.WriteVoid(w, err)
}

// --- Archive all (respects category filter) ---

func (h *Handler) ArchiveAll(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	category := r.URL.Query().Get("category")
	err := h.notifRepo.ArchiveAll(r.Context(), sessionCtx.Member.ID, category)
	httputil.WriteVoid(w, err)
}

// --- Soft delete ---

func (h *Handler) SoftDelete(w http.ResponseWriter, r *http.Request) {
	_, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.notifRepo.SoftDelete(r.Context(), id)
	httputil.WriteVoid(w, err)
}

// --- Undelete ---

func (h *Handler) Undelete(w http.ResponseWriter, r *http.Request) {
	_, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}

	err = h.notifRepo.Undelete(r.Context(), id)
	httputil.WriteVoid(w, err)
}

// --- Capabilities ---

type capabilitiesResponse struct {
	Channels        []string `json:"channels"`
	EmailConfigured bool     `json:"emailConfigured"`
	SMSConfigured   bool     `json:"smsConfigured"`
	InAppConfigured bool     `json:"inAppConfigured"`
}

func (h *Handler) Capabilities(w http.ResponseWriter, r *http.Request) {
	_, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	names := h.notifySvc.Registry().ConfiguredNames()
	if names == nil {
		names = []string{}
	}

	emailConfigured := false
	smsConfigured := false
	inAppConfigured := false
	for _, n := range names {
		switch n {
		case "email":
			emailConfigured = true
		case "sms":
			smsConfigured = true
		case "in_app":
			inAppConfigured = true
		}
	}

	httputil.WriteOK(w, capabilitiesResponse{
		Channels:        names,
		EmailConfigured: emailConfigured,
		SMSConfigured:   smsConfigured,
		InAppConfigured: inAppConfigured,
	})
}

// --- SSE Stream ---

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	rc := http.NewResponseController(w)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, unsubscribe := h.notifySvc.Hub().Subscribe(sessionCtx.Member.ID)
	defer unsubscribe()

	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	_ = rc.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			_ = rc.Flush()
		}
	}
}
