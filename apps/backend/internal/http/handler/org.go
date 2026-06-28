package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/pkg"
)

type OrgHandler struct {
	service    domainorg.Service
	sessionSvc session.Service
	cfg        config.Config
}

func NewOrgHandler(service domainorg.Service, sessionSvc session.Service, cfg config.Config) *OrgHandler {
	return &OrgHandler{service: service, sessionSvc: sessionSvc, cfg: cfg}
}

func (h *OrgHandler) DataSourceStatus(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetDataSourceStatus())
}

func (h *OrgHandler) DataSourceTest(w http.ResponseWriter, r *http.Request) {
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

func (h *OrgHandler) DataSourceUpdate(w http.ResponseWriter, r *http.Request) {
	cred, err := decodeCredential(r)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	force := r.URL.Query().Get("force") == "true"
	err = h.service.UpdateDataSource(r.Context(), cred, force)
	httputil.WriteVoid(w, err)
}

func (h *OrgHandler) DataSourceSearch(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	result, err := h.service.SearchDataSource(r.Context(), keyword)
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *OrgHandler) DataSourceImport(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ImportDataSource(r.Context())
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *OrgHandler) DataSourceImportRetry(w http.ResponseWriter, r *http.Request) {
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

func decodeCredential(r *http.Request) (domainorg.Credential, error) {
	cred, err := types.DecodeCredential(json.NewDecoder(r.Body))
	if err != nil {
		return domainorg.Credential{}, domain.BadRequest(httputil.MsgBadBody)
	}
	return cred, nil
}

func (h *OrgHandler) SyncConfigGet(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetSyncConfig())
}

func (h *OrgHandler) SyncConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var body domainorg.SyncConfig
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	h.service.UpdateSyncConfig(body)
	httputil.WriteVoid(w, nil)
}

func (h *OrgHandler) SyncTrigger(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.TriggerSync(r.Context())
	httputil.WriteJSON(w, http.StatusOK, result, err)
}

func (h *OrgHandler) SyncLogs(w http.ResponseWriter, r *http.Request) {
	page := pkg.ParseIntParam(r.URL.Query().Get("page"), 1)
	pageSize := pkg.ParseIntParam(r.URL.Query().Get("pageSize"), 10)
	httputil.WriteOK(w, h.service.ListSyncLogs(page, pageSize))
}

func (h *OrgHandler) DepartmentTree(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.GetDepartmentTree())
}

func (h *OrgHandler) DepartmentCreate(w http.ResponseWriter, r *http.Request) {
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

func (h *OrgHandler) DepartmentUpdate(w http.ResponseWriter, r *http.Request) {
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

func (h *OrgHandler) DepartmentDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.service.DeleteDepartment(r.Context(), id)
	httputil.WriteVoid(w, err)
}

func (h *OrgHandler) MembersList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := pkg.ParseIntParam(query.Get("page"), 1)
	pageSize := pkg.ParseIntParam(query.Get("pageSize"), 20)
	directOnly := query.Get("directOnly") == "true"
	httputil.WriteOK(w, h.service.ListMembers(
		query.Get("departmentId"), query.Get("keyword"), directOnly, page, pageSize,
	))
}

func (h *OrgHandler) MemberCreate(w http.ResponseWriter, r *http.Request) {
	var body domainorg.Member
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	member, err := h.service.CreateMember(body)
	httputil.WriteJSON(w, http.StatusOK, member, err)
}

func (h *OrgHandler) MemberUpdate(w http.ResponseWriter, r *http.Request) {
	var body domainorg.Member
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	member, err := h.service.UpdateMember(id, body)
	httputil.WriteJSON(w, http.StatusOK, member, err)
}

func (h *OrgHandler) MembersDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	err := h.service.DeleteMembers(r.Context(), body.IDs)
	httputil.WriteVoid(w, err)
}

func (h *OrgHandler) MembersStatus(w http.ResponseWriter, r *http.Request) {
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

func (h *OrgHandler) MembersTransfer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs          []string `json:"ids"`
		DepartmentID string   `json:"departmentId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	err := h.service.TransferMembers(r.Context(), body.IDs, body.DepartmentID)
	httputil.WriteVoid(w, err)
}

func (h *OrgHandler) MembersInvite(w http.ResponseWriter, r *http.Request) {
	_ = json.NewDecoder(r.Body).Decode(&struct{}{})
	err := h.service.InviteMember()
	httputil.WriteVoid(w, err)
}

func (h *OrgHandler) MembersBatchInvite(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	httputil.WriteOK(w, h.service.BatchInvite(body.IDs))
}

func (h *OrgHandler) MembersBatchImport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Rows []domainorg.BatchImportRow `json:"rows"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteOK(w, h.service.BatchImport(body.Rows))
}

func (h *OrgHandler) RolesList(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.ListRoles())
}

func (h *OrgHandler) RoleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteOK(w, h.service.CreateRole(body.Name, body.Permissions))
}

func (h *OrgHandler) RoleUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	role, err := h.service.UpdateRole(id, body.Name, body.Permissions)
	httputil.WriteJSON(w, http.StatusOK, role, err)
}

func (h *OrgHandler) RoleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.service.DeleteRole(id)
	httputil.WriteVoid(w, err)
}

func (h *OrgHandler) RoleMembersList(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	httputil.WriteOK(w, h.service.ListRoleMembers(roleID))
}

func (h *OrgHandler) RoleMemberAdd(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberID string `json:"memberId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	roleID := chi.URLParam(r, "roleId")
	err := h.service.AddRoleMember(roleID, body.MemberID)
	httputil.WriteVoid(w, err)
}

func (h *OrgHandler) RoleMemberRemove(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	memberID := chi.URLParam(r, "memberId")
	err := h.service.RemoveRoleMember(roleID, memberID)
	httputil.WriteVoid(w, err)
}

func (h *OrgHandler) PermissionsList(w http.ResponseWriter, r *http.Request) {
	httputil.WriteOK(w, h.service.ListPermissions())
}

func (h *OrgHandler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.PublicOrReadRoutes(h.cfg, r, h.sessionSvc, permission.OrgRead)
	read.Get("/data-source/status", h.DataSourceStatus)
	read.Get("/data-source/search", h.DataSourceSearch)
	read.Get("/sync/config", h.SyncConfigGet)
	read.Get("/sync/logs", h.SyncLogs)
	read.Get("/departments/tree", h.DepartmentTree)
	read.Get("/members", h.MembersList)
	read.Get("/roles", h.RolesList)
	read.Get("/roles/{roleId}/members", h.RoleMembersList)
	read.Get("/permissions", h.PermissionsList)

	write := httpmiddleware.WriteRoutes(r, h.cfg, h.sessionSvc)

	datasourceWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgDatasource))
	datasourceWrite.Post("/data-source/test", h.DataSourceTest)
	datasourceWrite.Put("/data-source", h.DataSourceUpdate)
	datasourceWrite.Post("/data-source/import", h.DataSourceImport)
	datasourceWrite.Post("/data-source/import/retry", h.DataSourceImportRetry)
	datasourceWrite.Put("/sync/config", h.SyncConfigUpdate)

	r.With(httpmiddleware.AllowSyncTrigger(h.cfg, h.sessionSvc)).Post("/sync/trigger", h.SyncTrigger)

	structureWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgStructure))
	structureWrite.Post("/departments", h.DepartmentCreate)
	structureWrite.Put("/departments/{id}", h.DepartmentUpdate)
	structureWrite.Delete("/departments/{id}", h.DepartmentDelete)

	membersWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgMembers))
	membersWrite.Post("/members", h.MemberCreate)
	membersWrite.Put("/members/{id}", h.MemberUpdate)
	membersWrite.Delete("/members", h.MembersDelete)
	membersWrite.Put("/members/status", h.MembersStatus)
	membersWrite.Post("/members/transfer", h.MembersTransfer)
	membersWrite.Post("/members/invite", h.MembersInvite)
	membersWrite.Post("/members/batch-invite", h.MembersBatchInvite)
	membersWrite.Post("/members/batch-import", h.MembersBatchImport)

	rolesWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgRoles))
	rolesWrite.Post("/roles", h.RoleCreate)
	rolesWrite.Put("/roles/{id}", h.RoleUpdate)
	rolesWrite.Delete("/roles/{id}", h.RoleDelete)
	rolesWrite.Post("/roles/{roleId}/members", h.RoleMemberAdd)
	rolesWrite.Delete("/roles/{roleId}/members/{memberId}", h.RoleMemberRemove)
}
