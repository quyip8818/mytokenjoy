package org

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
	var body struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	role, err := h.service.UpdateRole(r.Context(), id, body.Name, body.Permissions)
	httputil.WriteJSON(w, http.StatusOK, role, err)
}

func (h *Handler) RoleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.service.DeleteRole(r.Context(), id)
	httputil.WriteVoid(w, err)
}

func (h *Handler) RoleMembersList(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	members, err := h.service.ListRoleMembers(r.Context(), roleID)
	httputil.WriteJSON(w, http.StatusOK, members, err)
}

func (h *Handler) RoleMemberAdd(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberID string `json:"memberId"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	roleID := chi.URLParam(r, "roleId")
	err := h.service.AddRoleMember(r.Context(), roleID, body.MemberID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) RoleMemberRemove(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	memberID := chi.URLParam(r, "memberId")
	err := h.service.RemoveRoleMember(r.Context(), roleID, memberID)
	httputil.WriteVoid(w, err)
}

func (h *Handler) PermissionsList(w http.ResponseWriter, r *http.Request) {
	perms, err := h.service.ListPermissions(r.Context())
	httputil.WriteJSON(w, http.StatusOK, perms, err)
}
