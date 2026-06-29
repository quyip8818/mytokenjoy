package org

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (h *Handler) MembersList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := common.ParseIntParam(query.Get("page"), 1)
	pageSize := common.ParseIntParam(query.Get("pageSize"), 20)
	directOnly := query.Get("directOnly") == "true"
	httputil.WriteOK(w, h.service.ListMembers(
		query.Get("departmentId"), query.Get("keyword"), directOnly, page, pageSize,
	))
}

func (h *Handler) MemberCreate(w http.ResponseWriter, r *http.Request) {
	var body types.Member
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	member, err := h.service.CreateMember(body)
	httputil.WriteJSON(w, http.StatusOK, member, err)
}

func (h *Handler) MemberUpdate(w http.ResponseWriter, r *http.Request) {
	var body types.Member
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	member, err := h.service.UpdateMember(id, body)
	httputil.WriteJSON(w, http.StatusOK, member, err)
}

func (h *Handler) MembersDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.DeleteMembers(r.Context(), body.IDs)
	httputil.WriteVoid(w, err)
}

func (h *Handler) MembersStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs    []string `json:"ids"`
		Status string   `json:"status"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.UpdateMemberStatus(r.Context(), body.IDs, body.Status)
	httputil.WriteVoid(w, err)
}

func (h *Handler) MembersTransfer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs          []string `json:"ids"`
		DepartmentID string   `json:"departmentId"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.TransferMembers(r.Context(), body.IDs, body.DepartmentID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) MembersInvite(w http.ResponseWriter, r *http.Request) {
	var body struct{}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err := h.service.InviteMember()
	httputil.WriteVoid(w, err)
}

func (h *Handler) MembersBatchInvite(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteOK(w, h.service.BatchInvite(body.IDs))
}

func (h *Handler) MembersBatchImport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Rows []types.BatchImportRow `json:"rows"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteOK(w, h.service.BatchImport(body.Rows))
}
