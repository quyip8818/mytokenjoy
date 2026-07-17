package org

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
)

func (h *Handler) DepartmentTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.service.GetDepartmentTree(r.Context())
	httputil.WriteJSON(w, http.StatusOK, tree, err)
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
	parentID, _ := uuid.Parse(body.ParentID)
	dept, err := h.service.CreateDepartment(r.Context(), body.Name, parentID)
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
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	dept, err := h.service.UpdateDepartment(r.Context(), id, body.Name)
	httputil.WriteJSON(w, http.StatusOK, dept, err)
}

func (h *Handler) DepartmentDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	err = h.service.DeleteDepartment(r.Context(), id)
	httputil.WriteVoid(w, err)
}
