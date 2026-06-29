package org

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/http/httputil"
)

func (h *Handler) DataSourceStatus(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetDataSourceStatus())
}

func (h *Handler) DataSourceTest(w http.ResponseWriter, r *http.Request) {
	cred, err := decodeCredential(r)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	result, err := h.service.TestDataSource(r.Context(), cred)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	if !result.Success {
		httputil.WriteJSON(w, http.StatusUnprocessableEntity, result, nil)
		return
	}
	httputil.WriteOK(w, result)
}

func (h *Handler) DataSourceUpdate(w http.ResponseWriter, r *http.Request) {
	cred, err := decodeCredential(r)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	force := r.URL.Query().Get("force") == "true"
	err = h.service.UpdateDataSource(r.Context(), cred, force)
	httputil.WriteVoid(w, err)
}

func (h *Handler) DataSourceSearch(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	result, err := h.service.SearchDataSource(r.Context(), keyword)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) DataSourceImport(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ImportDataSource(r.Context())
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) DataSourceImportRetry(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	result, err := h.service.RetryImport(r.Context(), body.IDs)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}
