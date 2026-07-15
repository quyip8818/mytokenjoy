package notification

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
)

// --- List notifications ---

type notificationItemResponse struct {
	ID        string  `json:"id"`
	EventType string  `json:"eventType"`
	Channel   string  `json:"channel"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"createdAt"`
	ReadAt    *string `json:"readAt"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}

	entries, err := h.store.Notification().List(r.Context(), sessionCtx.Member.ID, limit, offset)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	items := make([]notificationItemResponse, len(entries))
	for i, e := range entries {
		items[i] = notificationItemResponse{
			ID:        e.ID,
			EventType: e.EventType,
			Channel:   e.Channel,
			Title:     e.Title,
			Body:      e.Body,
			Status:    e.Status,
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
		}
		if e.ReadAt != nil {
			s := e.ReadAt.Format(time.RFC3339)
			items[i].ReadAt = &s
		}
	}

	httputil.WriteOK(w, items)
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

	count, err := h.store.Notification().GetUnreadCount(r.Context(), sessionCtx.Member.ID)
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

	id := chi.URLParam(r, "id")
	if id == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "missing id")
		return
	}

	err := h.store.Notification().MarkRead(r.Context(), id)
	httputil.WriteVoid(w, err)
}

// --- Mark all read ---

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	err := h.store.Notification().MarkAllRead(r.Context(), sessionCtx.Member.ID)
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

	flusher, ok := w.(http.Flusher)
	if !ok {
		httputil.WriteStatus(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, unsubscribe := h.notifySvc.Hub().Subscribe(sessionCtx.Member.ID)
	defer unsubscribe()

	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

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
			flusher.Flush()
		}
	}
}
