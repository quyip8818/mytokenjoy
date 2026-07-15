package notification

import (
	"encoding/json"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
)

type preferenceEntryResponse struct {
	Category string `json:"category"`
	Channel  string `json:"channel"`
	Enabled  bool   `json:"enabled"`
}

type preferencesResponse struct {
	Preferences []preferenceEntryResponse `json:"preferences"`
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	entries, err := h.store.NotificationPreference().Get(r.Context(), sessionCtx.Member.ID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}

	items := make([]preferenceEntryResponse, len(entries))
	for i, e := range entries {
		items[i] = preferenceEntryResponse{
			Category: e.Category,
			Channel:  e.Channel,
			Enabled:  e.Enabled,
		}
	}

	httputil.WriteOK(w, preferencesResponse{Preferences: items})
}

type updatePreferencesRequest struct {
	Preferences []preferenceEntryResponse `json:"preferences"`
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	var req updatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Preferences) == 0 {
		httputil.WriteStatus(w, http.StatusBadRequest, "preferences required")
		return
	}

	entries := make([]types.NotificationPreferenceEntry, len(req.Preferences))
	for i, p := range req.Preferences {
		entries[i] = types.NotificationPreferenceEntry{
			Category: p.Category,
			Channel:  p.Channel,
			Enabled:  p.Enabled,
		}
	}

	err := h.store.NotificationPreference().Upsert(r.Context(), sessionCtx.Member.ID, entries)
	httputil.WriteVoid(w, err)
}

func (h *Handler) ResetPreferences(w http.ResponseWriter, r *http.Request) {
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	err := h.store.NotificationPreference().Delete(r.Context(), sessionCtx.Member.ID)
	httputil.WriteVoid(w, err)
}
