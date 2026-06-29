package org

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/http/httputil"
)

func (h *Handler) DepartmentTree(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetDepartmentTree())
}

func (h *Handler) DepartmentCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		ParentID string `json:"parentId"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	dept, err := h.service.CreateDepartment(r.Context(), body.Name, body.ParentID)
	httputil.WriteJSON(w, http.StatusOK, dept, err)
}

func (h *Handler) DepartmentUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	dept, err := h.service.UpdateDepartment(r.Context(), id, body.Name)
	httputil.WriteJSON(w, http.StatusOK, dept, err)
}

func (h *Handler) DepartmentDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.service.DeleteDepartment(r.Context(), id)
	httputil.WriteVoid(w, err)
}
