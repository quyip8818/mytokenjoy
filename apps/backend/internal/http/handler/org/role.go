package org

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/http/httputil"
)

func (h *Handler) RolesList(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context())
	httputil.WriteJSON(w, http.StatusOK, roles, err)
}

func (h *Handler) RoleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	role, err := h.service.CreateRole(r.Context(), body.Name, body.Permissions)
	httputil.WriteJSON(w, http.StatusOK, role, err)
}

func (h *Handler) RoleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	role, err := h.service.UpdateRole(r.Context(), id, body.Name, body.Permissions)
	httputil.WriteJSON(w, http.StatusOK, role, err)
}

func (h *Handler) RoleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid id")
		return
	}
	err = h.service.DeleteRole(r.Context(), id)
	httputil.WriteVoid(w, err)
}

func (h *Handler) RoleMembersList(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid roleId")
		return
	}
	members, err := h.service.ListRoleMembers(r.Context(), roleID)
	httputil.WriteJSON(w, http.StatusOK, members, err)
}

func (h *Handler) RoleMemberAdd(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid roleId")
		return
	}
	var body struct {
		MemberID uuid.UUID `json:"memberId"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	err = h.service.AddRoleMember(r.Context(), roleID, body.MemberID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) RoleMemberRemove(w http.ResponseWriter, r *http.Request) {
	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid roleId")
		return
	}
	memberID, err := uuid.Parse(chi.URLParam(r, "memberId"))
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid memberId")
		return
	}
	err = h.service.RemoveRoleMember(r.Context(), roleID, memberID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) PermissionsList(w http.ResponseWriter, r *http.Request) {
	perms, err := h.service.ListPermissions(r.Context())
	httputil.WriteJSON(w, http.StatusOK, perms, err)
}
