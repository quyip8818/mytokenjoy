package me

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
)

type loginActivityItem struct {
	Time      string `json:"time"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
	Current   bool   `json:"current"`
}

type loginActivityResponse struct {
	Items []loginActivityItem `json:"items"`
	Total int                 `json:"total"`
}

func (h *Handler) GetLoginActivity(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.SessionClaimsFromContext(r.Context())
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	sessions, total, err := h.sessions.ListByUser(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	items := make([]loginActivityItem, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, loginActivityItem{
			Time:      s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			IP:        s.IP,
			UserAgent: s.UserAgent,
			Current:   s.ID == claims.Sid,
		})
	}

	httputil.WriteOK(w, loginActivityResponse{Items: items, Total: total})
}
