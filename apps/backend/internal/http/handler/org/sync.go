package org

import (
	"net/http"

	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/pkg"
)

func (h *Handler) SyncConfigGet(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetSyncConfig())
}

func (h *Handler) SyncConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var body domainorg.SyncConfig
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.UpdateSyncConfig(body)
	httputil.WriteVoid(w, err)
}

func (h *Handler) SyncTrigger(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.TriggerSync(r.Context())
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) SyncLogs(w http.ResponseWriter, r *http.Request) {
	page := pkg.ParseIntParam(r.URL.Query().Get("page"), 1)
	pageSize := pkg.ParseIntParam(r.URL.Query().Get("pageSize"), 10)
	httputil.WriteOK(w, h.service.ListSyncLogs(page, pageSize))
}
