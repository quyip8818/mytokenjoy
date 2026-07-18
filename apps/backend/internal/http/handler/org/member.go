package org

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (h *Handler) MembersList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := common.ParseIntParam(query.Get("page"), 1)
	pageSize := common.ParseIntParam(query.Get("pageSize"), 20)
	directOnly := query.Get("directOnly") == "true"
	departmentID, _ := uuid.Parse(query.Get("departmentId"))
	result, err := h.service.ListMembers(r.Context(),
		departmentID, query.Get("keyword"), directOnly, page, pageSize,
	)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) MemberCreate(w http.ResponseWriter, r *http.Request) {
	var body types.Member
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	member, err := h.service.CreateMember(r.Context(), body)
	httputil.WriteJSON(w, http.StatusOK, member, err)
}

func (h *Handler) MemberUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body types.Member
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	member, err := h.service.UpdateMember(r.Context(), id, body)
	httputil.WriteJSON(w, http.StatusOK, member, err)
}

func (h *Handler) MembersDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []uuid.UUID `json:"ids"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context())
	if !ok {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	err := h.service.DeleteMembers(r.Context(), body.IDs, sessionCtx.Member.ID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) MembersStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs    []uuid.UUID `json:"ids"`
		Status string      `json:"status"`
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
		IDs          []uuid.UUID `json:"ids"`
		DepartmentID uuid.UUID   `json:"departmentId"`
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
		IDs []uuid.UUID `json:"ids"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	result, err := h.service.BatchInvite(r.Context(), body.IDs)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *Handler) MembersBatchImport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Rows []types.BatchImportRow `json:"rows"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	result, err := h.service.BatchImport(r.Context(), body.Rows)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}
