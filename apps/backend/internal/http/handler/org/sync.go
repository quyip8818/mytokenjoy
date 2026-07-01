package org

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (h *Handler) SyncConfigGet(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.service.GetSyncConfig(r.Context())
	httputil.WriteJSON(w, http.StatusOK, cfg, err)
}

func (h *Handler) SyncConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.SyncConfig
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.UpdateSyncConfig(r.Context(), body)
	httputil.WriteVoid(w, err)
}

func (h *Handler) SyncTrigger(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.TriggerSync(r.Context())
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) SyncLogs(w http.ResponseWriter, r *http.Request) {
	page := common.ParseIntParam(r.URL.Query().Get("page"), 1)
	pageSize := common.ParseIntParam(r.URL.Query().Get("pageSize"), 10)
	result, err := h.service.ListSyncLogs(r.Context(), page, pageSize)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}
